import abc
import json
import msgpack
import logging
from pydetector import detector
from traceback import format_exc
from python_driver.version import __version__
from typing import (Any, IO, NewType, Tuple, cast, List, Iterator, Dict)

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
Request    = NewType('Request',    Dict[str, Any])
Response   = NewType('Response',   Dict[AnyStr, Any])


class RequestCheckException(Exception):
    """
    Exception produced while there is an error during the processing
    of the request. It will cause an error reply to be produced on the
    output buffer.
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

    def _check_input_request(self, request: RawRequest) -> Tuple[str, str]:
        """
        Check the incoming request package and validate that the 'content' and
        'language' keys are not missing and that 'language' and 'language_version'
        are the right ones for this driver. It will also call _preprare_dict
        to covnvert the request bytestrings to str.

        :param request: The incoming request, already deserialized.

        .. raises::
            RequestCheckException if the request failed to validate.
        """
        str_request = self._tostr_request(request)
        code = asstr(str_request.get('content', ''))

        if not code:
            raise RequestCheckException('Bad input message, missing content')

        language = asstr(str_request.get('language', ''))
        if language.lower() != 'python':
            raise RequestCheckException('Bad language requested for the Python driver: "%s"' % language)

        language_version = asstr(str_request.get('language_version', ''))
        if language_version not in ('', '1', '2', '3'):
            raise RequestCheckException('Bad language version requested, Python driver only '
                                        'supports versions 1, 2 and 3')

        return code, asstr(str_request.get('filepath', ''))

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

    def _return_error(self, filepath: AnyStr='', status: AnyStr='error') -> None:
        """
        Build and send to stdout and error response. Also log
        the errors to the python_driver.log.

        :param filepath: optional str with the path of the source file that produced
        the error.
        :param status: error type, 'error' or 'fatal'
        """

        logging.error('Filepath: {}, Errors: {}'.format(filepath, self.errors))
        response = Response({
            'status': status,
            'errors': self.errors,
            'driver': 'python23:%s' % __version__,
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
            code, filepath = self._check_input_request(request)

            # We want the code detection to be fast and we prefer Python3 AST so using
            # the stop_on_ok_ast will avoid running a Python2 subprocess to check the
            # AST with Python2 if the Python3 version (which is a better AST anyway) worked
            resdict  = detector.detect(codestr=code, stop_on_ok_ast=True)
            codeinfo = resdict['<code_string>']
            version  = codeinfo['version']

            if version in (3, 6) and codeinfo['py3ast']:
                ast = codeinfo['py3ast']
            elif version in (1, 2) and codeinfo['py2ast']:
                ast = codeinfo['py2ast']
            else:
                raise Exception('Could not determine Python version')

            if not ast:
                raise Exception('Empty AST generated')

            response = Response({
                'status'           : 'ok',
                'errors'           : self.errors,
                'language'         : 'python',
                'language_version' : version,
                'driver'           : 'python23:%s' % __version__,
                'ast'              : ast,
            })
            if filepath:
                response['filepath'] = filepath

            self._send_response(response)

        except:
            status = 'fatal' if ast is None else 'error'
            self.errors.append(format_exc())
            self._return_error(filepath, status=status)

    @abc.abstractmethod
    def process_requests(self, inbuffer: InBuffer) -> None:
        """
        Main request-processing loop. It will call the _extract_docs iterator to
        get the requests. Must be reimplemented.

        :param inbuffer: file like object type str or abytes and supporting the readlines method.
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
        the readlines method.
        """
        for line in inbuffer.readlines():
            try:
                yield json.loads(line)
            except:
                self.errors = ['error decoding JSON from input: %s' % line]
                self._return_error(filepath='<jsonstream>', status='fatal')

    def process_requests(self, inbuffer: InStrBuffer) -> None:
        """
        Main request-processing loop. It will call the _extract_docs iterator to
        get the requests.

        :param inbuffer: file like object type str and supporting the readlines method.
        """
        for doc in self._extract_docs(inbuffer):
            self.process_request(doc)


class RequestProcessorMSGPack(RequestProcessor):
    """
    RequestProcessor subclass that operates deserializing and serializing responses
    using the MSGPACK format. Input and output packages
    """

    def __init__(self, outbuffer: OutBytesBuffer) -> None:
        """
        :param outbuffer: the output buffer. This must be a bytes-based file like object
        supporting the write(bytes) and flush() methods.
        """
        super().__init__(outbuffer)

    def _send_response(self, response: Response) -> None:
        self.outbuffer.write(msgpack.dumps(response))
        self.outbuffer.flush()

    def _tostr_request(self, request: RawRequest) -> Request:
        """
        Convert all byte-string keys and values to normal strings (non recursively since
        we only have one level)

        :param request: dictionary that potentially can contain bytestring keys and values.
        :return: the converted dictionary
        """
        try:
            newrequest = Request({})
            for key, value in request.items():
                if isinstance(value, bytes):
                    value = value.decode()

                if isinstance(key, bytes):
                    key = key.decode()

                newrequest[key] = value
        except AttributeError as exc:
            self.errors.append('Error trying to decode message, are you sure that the input ' +
                               'format is msgpack?')
            raise AttributeError from exc
        return newrequest

    def process_requests(self, inbuffer: InStrBuffer) -> None:
        """
        :param inbuffer: file-like object based on bytes supporting the read() and
        readlines() methods.
        """
        for request in msgpack.Unpacker(inbuffer):
            self.process_request(request)
