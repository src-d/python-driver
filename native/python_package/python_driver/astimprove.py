"""
Improve the basic AST, taken in dictionary form as exported
from the import pydetector module.
"""

import tokenize
import math
from codecs import encode
from copy import deepcopy
from io import BytesIO
from typing import List

from python_driver.base_token import Token, create_tokenized_lines
from python_driver.locationfixer import LocationFixer
from python_driver.noop_extractor import NoopExtractor
from python_driver.types import Node, AstDict, VisitResult

__all__ = ["AstImprover"]


class AstImprover():

    def __init__(self, codestr: str, astdict: AstDict) -> None:
        self._astdict = astdict
        # Tokenize and create the noop extractor and the position fixer
        self._tokens: List[Token] = [Token(*i) for i in tokenize.tokenize(BytesIO(codestr.encode('utf-8')).readline)]
        token_lines = create_tokenized_lines(codestr, self._tokens)
        self.noops_sync = NoopExtractor(codestr, token_lines)
        self.pos_sync   = LocationFixer(codestr, token_lines)
        self.codestr    = codestr

        # This will store a dict of nodes to end positions, it will be filled
        # on parse()
        self._node2endpos = None

        self.visit_Global = self.visit_Nonlocal = self._promote_names

    def parse(self) -> VisitResult:
        res = self.visit(self._astdict, root=True)
        return res

    def _normalize_position(self, node):
        # Python AST gives a 0 based column for the starting col, bblfsh uses 1-based
        if "col_offset" in node:
            node["col_offset"] = max(node.get("col_offset", 1) + 1, 1)

        if "end_col_offset" in node:
            node["end_col_offset"] = max(node["end_col_offset"] + 1, 1)

    def _remove_internal(self, node):
        node.pop('_fields', None)
        node.pop('_attributes', None)

    def visit(self, node: Node, root: bool=False) -> VisitResult:
        # the ctx property always has a "Load"/"Store"/etc dictionary that
        # can be perfectly converted to a string value since they don't
        # hold anything more than the name
        if isinstance(node, dict):
            node_type = node["ast_type"]
            if "ctx" in node:
                node["ctx"] = node["ctx"]["ast_type"]
        else:
            node_type = node.__class__.__name__

        meth = getattr(self, "visit_" + node_type, self.visit_other)
        visit_result = meth(node)
        self.noops_sync.add_noops(node, root)
        self.pos_sync.sync_node_pos(visit_result)

        if not self.codestr:
            # empty files are the only case where 0-indexes are allowed
            visit_result["col_offset"] = visit_result["end_col_offset"] = \
                    visit_result["lineno"] = visit_result["end_lineno"] = 0
        else:
            self._normalize_position(visit_result)

        self._remove_internal(visit_result)
        return visit_result

    def visit_str(self, node: Node) -> str:
        """
        This visits str fields inside nodes (which are represented as keys
        in the node dictionary), not Str AST nodes
        """
        return str(node)

    def visit_Bytes(self, node: Node) -> VisitResult:
        try:
            s = node["s"].decode()
            encoding = 'utf8'
        except UnicodeDecodeError:
            # try with base64
            s = encode(node["s"], 'base64').decode().strip()
            encoding = 'base64'

        node.update({"s": s, "encoding": encoding})
        return node

    def _promote_names(self, node: Node) -> VisitResult:
        # Python AST by default stores global and nonlocal variable names
        # in a "names" array of strings. That breaks the structure of everything
        # else in the AST (dictionaries, properties or list of objects) so we
        # convert those names to Name objects
        names_as_nodes = [self.visit({"ast_type": "Name",
                                      "id": i,
                                      "lineno": node["lineno"]})
                          for i in node["names"]]

        node["names"] = names_as_nodes
        return node

    def visit_NameConstant(self, node: Node) -> Node:
        if "value" in node:
            repr_val = repr(node["value"])
            if repr_val in ("True", "False"):
                node.update({"LiteralValue": "True" if node["value"] else "False",
                             "ast_type": "BoolLiteral"})
            elif repr_val == "None":
                node = self.visit_NoneType(node)
        else:
            node["ast_type"] = "NameConstant"
        return node

    def visit_Num(self, node: Node) -> Node:
        # complex objects are not json-serializable
        if isinstance(node["n"], complex):
            node.update({"n": {"real": node["n"].real,
                               "imag": node["n"].imag}})
        # infinity and nan are not json-serializable
        elif not math.isfinite(node["n"]):
            node.update({"n": str(node["n"])})
        return node

    def visit_NoneType(self, node: Node) -> Node:
        ret = node if node else {}
        ret.update({"LiteralValue": "None",
                    "ast_type": "NoneLiteral"})
        return ret

    def visit_Attribute(self, node: Node) -> Node:
        value = deepcopy(node.get("value"))

        if not value:
            return node

        ids: List[Node] = []

        while value:
            new_value = deepcopy(value.get("value"))
            if new_value:
                value.pop("value", None)
            ids.insert(0, self.visit(value))
            value = new_value

        # Append a copy of this node at the end, and change the type of the original
        node.pop("value", None)
        node_copy = deepcopy(node)
        self._remove_internal(node_copy)
        self._normalize_position(node_copy)
        ids.append(node_copy)

        node["ast_type"] = "QualifiedIdentifier"

        # Copy the position of the first element
        if len(ids):
            for key in ("lineno", "end_lineno", "col_offset", "end_col_offset"):
                if ids[0].get(key):
                    node[key] = ids[0][key]

        node.pop("attr", None)
        node["identifiers"] = ids

        return node

    def visit_arguments(self, node: Node) -> Node:
        """
        Convert the very odd Python's argument node organization (several different lists
        for each type and each type's default arguments that you have to right-match) into
        a more common in other languages single list of types arguments with default
        values as children of their arg. Also convert Python2's "Name" types inside the
        arguments to
        """

        def match_default_args(args: List[Node], defaults: List[Node]) -> List[Node]:
            if defaults:
                lendiff = len(args) - len(defaults)

                for i, arg in enumerate(args[lendiff:]):
                    arg["default"] = self.visit(defaults[i])

            return args

        def name2arg(node: Node):
            # Normalize Python2 and 3 argument types
            if node["ast_type"] == "Name":
                node["ast_type"] = "arg"
            id_ = node.get("id")
            if id_:
                node["@token"] = node["id"]
                del node["id"]

        norm_args: List[Node] = []

        normal_args = deepcopy(node.get("args"))
        if normal_args:
            defaults = deepcopy(node.get("defaults"))
            match_default_args(normal_args, defaults)

            for i in normal_args:
                norm_args.append(self.visit(i))

        kwonly_args = deepcopy(node.get("kwonlyargs"))
        if kwonly_args:
            kw_defaults = deepcopy(node.get("kw_defaults"))
            match_default_args(kwonly_args, kw_defaults)

            for a in kwonly_args:
                a["ast_type"] = "kwonly_arg"

            for i in kwonly_args:
                norm_args.append(self.visit(i))

        kwarg = deepcopy(node.get("kwarg"))
        if kwarg:
            if isinstance(kwarg, str):
                # Python2 kwargs are just strings; convert to same format
                # as Python3
                kwarg = {
                    "arg": kwarg,
                    "annotation": None,
                    # the tokenizer will fix the positions later
                    "lineno": 1,
                    "end_lineno": 1,
                    "col_offset": 0,
                    "end_col_offset": 0
                    }
            kwarg["ast_type"] = "kwarg"
            norm_args.append(self.visit(kwarg))

        vararg = deepcopy(node.get("vararg"))
        if vararg:
            vararg["ast_type"] = "vararg"
            norm_args.append(self.visit(vararg))

        for k in ('defaults', 'kw_defaults', 'args', 'kwonlyargs', 'kwarg', 'vararg'):
            if k in node:
                del node[k]

        for n in norm_args:
            if "arg" in n:
                n["@token"] = n["arg"]
                del n["arg"]
            name2arg(n)

        node["args"] = norm_args
        return node


    def visit_other(self, node: Node) -> VisitResult:
        for field in node.get("_fields", []):
            meth = getattr(self, "visit_" + node["ast_type"], self.visit_other_field)
            child = node.get(field)
            if child:
                node[field] = meth(child)
        return node

    def visit_other_field(self, node: Node) -> VisitResult:
        if isinstance(node, dict):
            return self.visit(node)
        elif isinstance(node, list) or isinstance(node, tuple):
            return [self.visit(x) for x in node]
        else:
            # string attribute
            return node


if __name__ == '__main__':
    import sys
    import importlib.util
    from pprint import pprint

    if len(sys.argv) > 1:
        from pydetector import detector
        codestr = open(sys.argv[1]).read()
        resdict = detector.detect(codestr=codestr, stop_on_ok_ast=True)
        codeinfo = resdict['<code_string>']
        version = codeinfo['version']

        failed = False
        testdict = None

        if version in (3, 6) and codeinfo['py3ast']:
            testdict = codeinfo['py3ast']["PY3AST"]
            print("Using Python3")
        elif version in (1, 2) and codeinfo['py2ast']:
            testdict = codeinfo['py2ast']["PY2AST"]
            print("Using Python2")
        else:
            failed = True
            errors = [
              'Errors produced trying to get an AST for both Python versions' +
              '\n------ Python2 errors:\n%s' % codeinfo['py2_ast_errors'] +
              '\n------ Python3 errors:\n%s' % codeinfo['py3_ast_errors']
            ]

        if not failed and not testdict:
            raise Exception('Empty AST generated from non empty code')
        ast = AstImprover(codestr, testdict).parse()
        if not ast:
            raise Exception('Empty AST generated from non empty code')
    else:
        codestr = open("../test/fixtures/detector.py").read()
        spec = importlib.util.spec_from_file_location(
                "module.testmod",
                "../test/fixtures/exported_dict.py"
        )
        testmod = importlib.util.module_from_spec(spec)

        if spec.loader:
            spec.loader.exec_module(testmod)
            testdict = testmod.testdict  # type: ignore

    pprint(AstImprover(codestr, testdict).parse())
