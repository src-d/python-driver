from __future__ import print_function

import ast
import tokenize
import token as token_module
from six import StringIO
from codecs import encode
from pprint import pprint

# TODO: add an option to not change the node names of NameConstant, Num and Str

def export_dict(codestr):
    visitor = DictExportVisitor(codestr)
    return visitor.parse()


def export_json(codestr, pretty_print=False):
    import json
    dict_ = export_dict(codestr)
    json_ = json.dumps(dict_, indent=2 if pretty_print else 0, ensure_ascii=False)
    return json_, dict_

namecounter = 0
def export_graphviz(codestr):
    """
    WARNING: Experimental, only compatible with Python 3 and unfinished
    """
    # TODO: use fillcolor for different ast_types
    # TODO: use shape for giving different shapes to different nodes: http://www.graphviz.org/content/node-shapes#polygon
    parent_nodekeys = {'body', 'names', 'targets', 'value', 'func', 'args', 'keywords',
                       'keys', 'left', 'right'}
    from graphviz import Graph

    # These nodes will show their childs as hanging from their parent
    def generate(dot, obj, parent=None, reparent=False):
        global namecounter

        namecounter += 1
        nodename = str(namecounter)

        def get_name(node):
            subname = ''
            if isinstance(node, dict):
                ast = node.get('ast_type', '')
                subname = node.get('name', '')
                id = node.get('id', '')
                arg = node.get('arg', '')
                separator = ': ' if ast and (subname or id or arg) else ''
                name = '%s%s%s%s%s' % (ast, separator, subname, id, arg)
            else:
                name = str(type(node))
            return name, subname

        name, origname = get_name(obj)
        if reparent or type(obj) in (list, tuple):
            nodename = parent
        else:
            dot.node(nodename, label=name)
            if parent:
                dot.edge(parent, nodename)

        if isinstance(obj, dict):
            for childkey in parent_nodekeys:
                if childkey in obj:
                    reparent = childkey == 'names'
                    generate(dot, obj[childkey], nodename, reparent)

        elif isinstance(obj, list) or isinstance(obj, tuple):
            for value in obj:
                generate(dot, value, nodename)

    dict_ = export_dict(codestr),
    dot = Graph(comment="Python AST", format='pdf',
                graph_attr={'ranksep': '2.5'
                            })
    generate(dot, dict_, 'RootNode')

    with open('ast.dot', 'w') as dotfile:
        print(dot.source, file=dotfile)


TOKEN_VALUE = 1
TOKEN_RAWVALUE = 4
NOOP_TOKENS_LINE = {'COMMENT', 'INDENT', 'NL', 'NEWLINE'}

class NoopExtractor(object):
    """
    Tokenize the source code and extract lines with tokens that Python's
    AST generator ignore like blanks and comments
    """

    def __init__(self, codestr):
        tokens = tokenize.generate_tokens(StringIO(codestr).readline)

        self.current_line = None
        self.all_lines = self._create_tokenized_lines(tokens, codestr)
        self.astmissing_lines = self._create_astmissing_lines()

    def _create_astmissing_lines(self):
        """
        Return a copy of line_tokens containing lines ignored by the AST
        (comments and blanks-only lines)
        """
        lines = []
        nl_token = (token_module.NEWLINE, '\n', (0, 0), (0, 0), '\n')

        ltname = token_module.tok_name
        for i, linetokens in enumerate(self.all_lines):
            if len(linetokens) == 1 and ltname[linetokens[0][0]] == 'NL':
                lines.append(nl_token)
            else:
                for token in linetokens:
                    if ltname[token[0]] == 'COMMENT' and \
                            token[TOKEN_RAWVALUE].lstrip().startswith('#'):
                        lines.append(token)
                        break
                else:
                    lines.append(None)
        assert len(lines) == len(self.all_lines)

        for i, linetokens in enumerate(lines):
            if linetokens:
                self.current_line = i
                break
        else:
            self.current_line = len(lines)
        return lines

    def _create_tokenized_lines(self, tokens, codestr):
        """
        Create a list of tokenized lines
        """
        lines = codestr.splitlines() if codestr else []
        result = []
        for i in range(0, len(lines) + 1):
            result.append([])

        ltname = token_module.tok_name
        for token in tokens:
            tokentype = ltname[token[0]]
            srow, scol = token[2]
            erow, ecol = token[3]
            line = erow - 1 if tokentype == 'STRING' else srow - 1
            result[line].append(token)
        assert len(lines) + 1 == len(result), len(result)
        return result

    def previous_nooplines(self, node):
        """Return a list of the preceding comment and blank lines"""
        previous = []
        if hasattr(node, 'lineno'):
            while self.current_line < node.lineno:
                token = self.astmissing_lines[self.current_line]
                if token:
                    s = token[TOKEN_RAWVALUE].rstrip() + '\n'
                    previous.append(s)
                self.current_line += 1
        return previous

    def remainder_noops_sameline(self, node):
        """
        Return a string containing the trailing (until EOL) noops for the
        node, if any. The ending newline is implicit and thus not returned
        """
        if not hasattr(node, 'lineno'):
            return ''

        tokens = self.all_lines[node.lineno - 1]
        trailing = []
        ltname = token_module.tok_name

        for token in tokens:
            if ltname[token[0]] not in NOOP_TOKENS_LINE:
                # restart
                trailing = []
            else:
                trailing.append(token[TOKEN_VALUE])
        return ''.join(trailing[:-1])

    def remainder_noops(self):
        """return any remaining ignored lines."""
        trailing = []
        i = self.current_line
        while i < len(self.astmissing_lines):
            token = self.astmissing_lines[i]
            if token:
                s = token[TOKEN_RAWVALUE]
                trailing.append(s)
            i += 1
        self.current_line = i
        return trailing

def node_dict(node, newdict, ast_type=None):
    """
    Shortcut that adds ast_type (if not specified),
    lineno and col_offset to the node-derived dictionary
    """
    if ast_type is None:
        ast_type = node.__class__.__name__

    newdict["ast_type"] = ast_type
    if hasattr(node, "lineno"):
        newdict["lineno"] = node.lineno
    if hasattr(node, "col_offset"):
        newdict["col_offset"] = node.col_offset

    return newdict

class DictExportVisitor(object):
    ast_type_field = "ast_type"

    def __init__(self, codestr, ast_parser=ast.parse, tsync_class=NoopExtractor):
        """
        Initialize the Token Syncer composited object, parse the source code
        and start visiting the node tree to add comments and other modifications

        Args:
            codestr (string): the string with the source code to parse and visit.

            ast_parser (function, optional): the AST parser function to use. It needs to take
            a string with the code as parameter. By default it will be ast.parse from stdlib.

            tsync_class (class, optional): the class to use to sinchronize the tokenizer with
            the AST visits. This is needed to extract aditional info like comments or whitespace
            that most AST parsers doesn't include in the tree. This class need to provide the public
            methods "previous_nooplines",  "remainder_noops_sameline" and "rmainder_noops".
            By default astexport.NoopExtractor will be used.
        """
        self.codestr = codestr
        self.sync = tsync_class(codestr)

    def parse(self):
        node = ast.parse(self.codestr, mode='exec')
        res = self.visit(node, root=True)
        return res


    def visit(self, node, root=False):
        node_type = node.__class__.__name__
        # the ctx property always has a "Load"/"Store"/etc nodes that
        # can be perfectly converted to a string value since they don't
        # hold anything more than the name
        if hasattr(node, 'ctx'):
            node.ctx = node.ctx.__class__.__name__

        meth = getattr(self, "visit_" + node_type, self.visit_other)
        visit_result = meth(node)

        if isinstance(visit_result, dict):
            # Add all the noop (whitespace and comments) lines between the
            # last node and this one
            noops_previous = self.sync.previous_nooplines(node)
            if noops_previous:
                visit_result['noops_previous'] = {
                    "ast_type": "PreviousNoops",
                    "lines": [{"ast_type": "NoopLine", "l": noopline} for noopline in noops_previous]
                }

            # Other noops at the end of its significative line except the implicit
            # finishing newline
            noops_sameline = self.sync.remainder_noops_sameline(node)
            if noops_sameline:
                visit_result['noops_sameline'] = {
                    "ast_type": "SameLineNoops",
                    "lines": [{"ast_type": "NoopLine", "l": noopline} for noopline in noops_sameline]
                }

            # Finally, if this is the root node, add all noops after the last op node
            if root:
                noops_remainder = self.sync.remainder_noops()
                if noops_remainder:
                    visit_result['noops_remainder'] = {
                        "ast_type": "RemainderNoops",
                        "lines": [{"ast_type": "NoopLine", "l": noopline} for noopline in noops_remainder]
                    }
        return visit_result

    def visit_other(self, node):
        node_type = node.__class__.__name__

        node_dict = {
            self.ast_type_field: node_type
        }

        # Visit fields
        for field in node._fields:
            meth = getattr(
                self, "visit_" + node_type,
                self.visit_other_field
            )
            node_dict[field] = meth(getattr(node, field))

        # Visit attributes
        for attr in node._attributes:
            meth = getattr(
                self, "visit_" + node_type + "_" + attr,
                self.visit_other_field
            )
            # Use None as default when lineno/col_offset are not set
            node_dict[attr] = meth(getattr(node, attr, None))
        return node_dict

    def visit_other_field(self, node):
        if isinstance(node, ast.AST):
            return self.visit(node)
        elif isinstance(node, list) or isinstance(node, tuple):
            return [self.visit(x) for x in node]
        else:
            return node

    def visit_str(self, node):
        """
        This visits str fields inside nodes (which are represented as keys
        in the node dictionary), not Str AST nodes
        """
        return str(node)

    def visit_Str(self, node):
        return {
            self.ast_type_field: "StringLiteral",
            "LiteralValue": node.s
        }

    def visit_Bytes(self, node):
        try:
            s = node.s.decode()
            encoding = 'utf8'
        except UnicodeDecodeError:
            # try with base64
            s = encode(node.s, 'base64').decode().strip()
            encoding = 'base64'

        return {
            self.ast_type_field: "ByteLiteral",
            "LiteralValue": s,
            "encoding": encoding,
        }

    def visit_NoneType(self, node):
        return 'NoneLiteral'

    def visit_Global(self, node):
        # Python AST by default stores global and nonlocal variable names
        # in a "names" array of strings. That breaks the structure of everything
        # else in the AST (dictionaries, properties or list of objects) so we
        # convert those names to Name objects

        names_as_nodes = [{"ast_type": "Name",
                          "id": i,
                          "lineno": node.lineno,
                          "col_offset": node.col_offset} for i in node.names]
        return node_dict(node, {"names": names_as_nodes}, ast_type="Global")

    def visit_Nonlocal(self, node):
        # ditto
        names_as_nodes = [{"ast_type": "Name",
                          "id": i,
                          "lineno": node.lineno,
                          "col_offset": node.col_offset} for i in node.names]
        return node_dict(node, {"names": names_as_nodes}, ast_type="Nonlocal")


    def visit_NameConstant(self, node):
        if hasattr(node, 'value'):
            repr_val = repr(node.value)
            if repr_val in ('True', 'False'):
                return {
                    self.ast_type_field: "BoolLiteral",
                    "LiteralValue": node.value
                }
            elif repr_val == 'None':
                return {
                    self.ast_type_field: "NoneLiteral",
                    "LiteralValue": node.value
                }
        return str(node)

    def visit_Num(self, node):
        if isinstance(node.n, int):
            retDict = {
                    "NumType": "int",
                    "LiteralValue": node.n
                    }
        elif isinstance(node.n, float):
            retDict = {
                    "NumType": "float",
                    "LiteralValue": node.n
                    }
        elif isinstance(node.n, complex):
            retDict = {
                    "NumType": "complex",
                    "LiteralValue": {"real": node.n.real, "imaginary": node.n.imag},
                    }

        return node_dict(node, retDict, ast_type="NumLiteral")


if __name__ == '__main__':
    import sys

    f = sys.argv[1]

    with open(f) as codefile:
        content = codefile.read()
    # pprint(export_dict(content))
    print(export_json(content, pretty_print=True)[0])
    # export_graphviz(content)
