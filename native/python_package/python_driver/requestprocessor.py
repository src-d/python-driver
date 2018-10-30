import abc
import json
import sys
try:
    from pydetector import detector
except ImportError:
    # local test mode
    sys.path.insert(0, '../../.local/lib/python3.6/site-packages')
    from pydetector import detector

from traceback import format_exc
from python_driver.version import __version__
from python_driver.astimprove import AstImprover, VisitResult
from typing import (Any, IO, NewType, cast, List, Iterator, Dict, Optional)

# typing.AnyStr is bugged on this version of MyPy, so...:
AnyStr = Any

# aliases
InBuffer       = IO[AnyStr]
OutBuffer      = IO[AnyStr]
OutStrBuffer   = IO[str]
InStrBuffer    = IO[str]
OutBytesBuffer = IO[bytes]
InBytesBuffer  = IO[bytes]

# types
RawRequest = NewType('RawRequest', Dict[AnyStr, Any])
Request    = NewType('Request', Dict[str, Any])
Response   = NewType('Response', Dict[AnyStr, Any])


class EmptyCodeException(Exception):
    """
    Exception produced when the input code is empty. This should
    generate an error (because we can't parse anything) but not
    a fatal one (e. g. empty __init__.py files)
    """
    pass


def asstr(txt: AnyStr) -> str:
    """
    Convert from byte to str. Noop if it was already an str
    """
    if isinstance(txt, bytes):
        return txt.decode()
    return txt


class RequestProcessor(metaclass=abc.ABCMeta):
    """
    Base class of RequestProcessors. This must be subclassed to add new communication
    protocols. This should be done reimplementing the methods: _send_response,
    _tostr_request and process_requests.
    """
    def __init__(self, outbuffer: OutBuffer) -> None:
        """
        Constructor for a RequestProcessor instance.

        :param outbuffer: the output buffer. This must be a file-like object based on
        str or bytes depnding on the protocol (see the documentation on subclasses for
        specifics) supporting the write(type) and flush() methods.
        """
        self.outbuffer = outbuffer
        self.errors: List[str] = []

    @abc.abstractmethod
    def _tostr_request(self, request: RawRequest) -> Request:
        """
        This method should be reimplemented by subclasses ensuring that all the keys
        and values in the request dictionary once deserialized have the str type.
        :param request: the deserialized request dictionary
        :return: the converted dictionary
        """
        pass

    @abc.abstractmethod
    def _send_response(self, response: Response) -> None:
        """
        Send the response dictionary. This method must be reimplemented by
        protocol handling subclasses and must encapsulate the operations needed
        to serialize the response dictionary to the final format and send it
        on the output buffer

        :param response: the response dictionary, not serialized
        """
        pass

    def _return_error(self, filepath: AnyStr='', status: AnyStr='error',
            ast: Optional[VisitResult] = None) -> None:
        """
        Build and send to stdout and error response. Also log
        the errors to the python_driver.log.

        :param filepath: optional str with the path of the source file that produced
        the error.
        :param status: error type, 'error' or 'fatal'
        """

        ret_ast = None if status == 'fatal' else ast
        response = Response({
            'status': status,
            'errors': self.errors,
            'driver': 'python23:%s' % __version__,
            'ast': ret_ast,
        })
        if filepath:
            response['filepath'] = filepath
        self._send_response(response)

    def process_request(self, request: RawRequest) -> None:
        """
        Main function doing the work of processing a single request. It will
        do its best effort to detect the code Python version(s), extract the AST,
        prepare the reply package and sent it on the output buffer.

        Any error will cause an error or fatal package to be submitted.

        :param request: the deserialized request dictionary.
        """
        filepath    = ''
        ast         = None
        self.errors = []

        try:
            str_request = self._tostr_request(request)
            code = asstr(str_request.get('content', ''))

            if code:
                # We want the code detection to be fast and we prefer Python3 AST so using
                # the stop_on_ok_ast will avoid running a Python2 subprocess to check the
                # AST with Python2 if the Python3 version (which is a better AST anyway) worked
                resdict  = detector.detect(codestr=code, stop_on_ok_ast=True)
                codeinfo = resdict['<code_string>']
                version  = codeinfo['version']

                failed = False

                if version in (3, 6) and codeinfo['py3ast']:
                    orig_ast = codeinfo['py3ast']["PY3AST"]
                elif version in (1, 2) and codeinfo['py2ast']:
                    orig_ast = codeinfo['py2ast']["PY2AST"]
                else:
                    failed = True
                    self.errors = [
                        'Errors produced trying to get an AST for both Python versions' +
                        '\n------ Python2 errors:\n%s' % codeinfo['py2_ast_errors'] +
                        '\n------ Python3 errors:\n%s' % codeinfo['py3_ast_errors']
                    ]

                if not failed:
                    if not orig_ast:
                        raise Exception('Empty AST generated from non empty code')
                    ast = AstImprover(code, orig_ast).parse()
                    if not ast:
                        raise Exception('Empty AST generated from non empty code')
            else:
                # Module with empty code (like __init__.py) return a module-only AST
                # since this would still have semantic meaning for Python
                ast = {
                        "ast_type"   : "Module",
                        "lineno"     : 1,
                        "col_offset" : 1,
                       }
                version = 3

            response = Response({
                'errors'           : self.errors,
                'metadata'         : {
                    'language'         : 'python',
                    'language_version' : version,
                    'driver'           : 'python23:%s' % __version__,
                }
            })

            if failed:
                response['status'] = 'error'
                response['ast'] = None
            else:
                response['status'] = 'ok'
                response['ast'] = {"PY%dAST" % version: ast}

            if filepath:
                response['filepath'] = filepath

            self._send_response(response)

        except EmptyCodeException:
            self.errors.append('Code field empty')
            self._return_error(filepath, status='error', ast=ast)

        except:
            status = 'fatal' if ast is None else 'error'
            self.errors.append(format_exc())
            self._return_error(filepath, status=status)

    @abc.abstractmethod
    def process_requests(self, inbuffer: InBuffer) -> None:
        """
        Main request-processing loop. It will call the _extract_docs iterator to
        get the requests. Must be reimplemented.

        :param inbuffer: file like object type str or abytes and supporting the iteration by lines
        """
        pass


class RequestProcessorJSON(RequestProcessor):
    """
    RequestProcessor subclass that operates deserializing requests and serializing
    responses using the JSON format. When sending JSON documents on the inputbuffer
    the documents must be separated with a Unix newline ('\n') .
    The documents produced on the output stream by this
    class will also have this newline separator between them.
    """

    def __init__(self, outbuffer: OutStrBuffer) -> None:
        """
        :param outbuffer: the output buffer. This must be a file-like object based on
        str (like sys.stdout or io.StringIO but not sys.stdout.buffer or io.BytesIO).
        """
        super().__init__(outbuffer)

    def _send_response(self, response: Response) -> None:
        """
        Serialized the response dictionary to JSON format and sent it on the
        (str-based) output buffer followed by the document separator line.

        The generated JSON content could have non-ascii characters.

        :param response: deserialized response dictionary
        """
        self.outbuffer.write(json.dumps(response, ensure_ascii=False))
        self.outbuffer.write('\n')
        self.outbuffer.flush()

    def _tostr_request(self, request: RawRequest) -> Request:
        """
        This interface method is a NOOP for this class since JSON keys already comes
        encoded as str so there is no need to lose time reencoding anything

        :return: the dictionary with str keys and values (where applicable) converted
        to str
        """
        return cast(Request, request)

    def _extract_docs(self, inbuffer: InStrBuffer) -> Iterator[RawRequest]:
        """
        This generator will read the inbuffer yielding the JSON
        docs one by every line, without using other separators than '\n'

        :param inbuffer: the input buffer that be of type str and support
        iteration by lines.
        """

        for line in inbuffer:
            try:
                yield json.loads(line)
            except:
                self.errors = ['error decoding JSON from input: %s' % line]
                self._return_error(filepath='<jsonstream>', status='fatal')

    def process_requests(self, inbuffer: InStrBuffer) -> None:
        """
        Main request-processing loop. It will call the _extract_docs iterator to
        get the requests.

        :param inbuffer: file like object type str and supporting iteration by lines.
        """
        for doc in self._extract_docs(inbuffer):
            self.process_request(doc)
