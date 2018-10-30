import io
import json
import os
import subprocess
import sys
import unittest
from os.path import join, abspath, dirname

sys.path.append('..')
from python_driver import __version__, get_processor_instance
from python_driver.requestprocessor import (
    Request, Response, RequestProcessorJSON, InBuffer, EmptyCodeException)

from typing import Dict, Any, List, AnyStr, Optional, Iterator, cast

CURDIR = abspath(dirname(__file__))

# Disabled until I update the new module with typing
# class TestTypeCheck(unittest.TestCase):
    # def test_10_check(self) -> None:
        # prevdir = os.getcwd()
        # try:
            # os.chdir(dirname(CURDIR))
            # srcdir = abspath(join(dirname(CURDIR), 'python_driver', '*'))
            # self.assertEqual(subprocess.call(['test/typecheck.sh', srcdir], shell=True), 0)
        # finally:
            # os.chdir(prevdir)


class TestPythonDriverBase(unittest.TestCase):
    def _restart_data(self, format_: str='json') -> None:
        assert format_ == 'json'

        with open(join(CURDIR, 'data', 'helloworld.py')) as f:
            testcode = f.read()

        self.data = Request({
            'filepath': 'test.py',
            'action': 'ParseAST',
            'content': testcode,
            'language': 'python',
        })

        bufferclass = io.StringIO if format_ == 'json' else io.BytesIO

        # This will mock the python_driver stdin
        self.sendbuffer = bufferclass()
        # This will mock the python_driver stdout
        self.recvbuffer = bufferclass()

    @staticmethod
    def _extract_docs(inbuffer: InBuffer) -> Iterator[Response]:
        """
        This generator will read the inbuffer yielding the JSON
        docs when it finds the ending mark
        """
        line: str
        for line in inbuffer.readlines():
            yield json.loads(line)

    def _loadResults(self, format_: str) -> List[Response]:
        """Read all msgs from the recvbuffer"""
        self.recvbuffer.seek(0)

        res: List[Response] = []
        res = [doc for doc in self._extract_docs(self.recvbuffer)]
        return res


class Test10ProcessRequestFunc(TestPythonDriverBase):

    def _add_to_buffer(self, count: int, format_: str) -> None:
        """Add count test msgs to the sendbuffer"""
        for i in range(count):
            msg = ''
            msg = json.dumps(self.data, ensure_ascii=False) + '\n'
            self.sendbuffer.write(msg)

        self.sendbuffer.flush()

    def _send_receive(self, nummsgs: int, outformat: str='json',
                      dataupdate: Optional[Dict[AnyStr, Any]]=None,
                      restart_data: bool=True) -> List[Response]:
        if restart_data:
            self._restart_data(outformat)

        if dataupdate:
            self.data.update(dataupdate)

        self._add_to_buffer(nummsgs, outformat)
        self.sendbuffer.seek(0)

        processor, _ = get_processor_instance(
                outformat,
                custom_outbuffer=self.recvbuffer,
                custom_inbuffer=self.sendbuffer
        )
        processor.process_requests(self.sendbuffer)
        return self._loadResults(outformat)

    def _check_reply_dict(self, response: Response, has_errors: bool=False) -> None:
        self.assertIsInstance(response, dict)
        status = response.get('status')

        if has_errors:
            assert status in ('error', 'fatal')
            errors = response.get('errors', list)
            self.assertIsInstance(errors, list)
            self.assertGreater(len(errors), 0)
        else:
            self.assertEqual(status, 'ok')
            self._check_AST_dict(response)
            language_version = response['metadata'].get('language_version', -1)
            assert str(language_version) in ('2', '3')

    def _check_AST_dict(self, response: Response) -> None:
        self.assertIsNotNone(response)
        assert 'ast' in response
        self.assertIsInstance(response['ast'], dict)
        root_key = list(response['ast'].keys())[0]
        assert root_key

        for key in ('ast_type', 'body'):
            assert key in response['ast'][root_key]

        self.assertIsInstance(response['ast'][root_key]['body'], list)

        for item in response['ast'][root_key]['body']:
            for key in ('ast_type', 'lineno', 'col_offset'):
                assert key in item

    def test_010_normal_json(self) -> None:
        replies = self._send_receive(1, 'json')
        self.assertEqual(len(replies), 1)
        self._check_reply_dict(replies[0])

    def test_020_normal_json_many(self) -> None:
        replies = self._send_receive(100, 'json')
        self.assertEqual(len(replies), 100)
        for reply in replies:
            self._check_reply_dict(reply)

    def test_030_error_print(self) -> None:
        wrongcode = 'wtf lol'

        replies = self._send_receive(1, 'json', {'content': wrongcode})
        self.assertEqual(len(replies), 1)
        ast = replies[0].get('ast')
        self.assertIsNone(ast)
        self._check_reply_dict(replies[0], has_errors=True)

        # Check that it still alive
        self._restart_data()
        replies = self._send_receive(1, 'json')
        self.assertEqual(len(replies), 1)

    def test_040_broken_json(self) -> None:
        self._restart_data('json')
        brokendata = json.dumps(self.data, ensure_ascii=False)[:-30]
        self.sendbuffer.write(brokendata)
        self.sendbuffer.flush()
        reply = self._send_receive(1, 'json', restart_data=False)[0]
        self.assertEqual(reply['status'], 'fatal')
        self.assertEqual(len(reply['errors']), 1)


class Test20ReqProcMethods(TestPythonDriverBase):

    def test_10_send_response_json(self) -> None:
        self._restart_data('json')
        processor = RequestProcessorJSON(self.recvbuffer)
        processor._send_response(cast(Response, self.data))
        res = self._loadResults('json')
        self.assertEqual(len(res), 1)
        self.assertDictEqual(self.data, res[0])

    # process request already tested with TestPythonDriverBase

    def test_20_return_error(self) -> None:
        self._restart_data('json')
        processor = RequestProcessorJSON(self.recvbuffer)
        processor.errors = ['test error']
        processor._return_error('test.py', 'fatal')
        res = self._loadResults('json')
        self.assertEqual(len(res), 1)
        self.assertDictEqual(res[0] , {'driver': 'python23:%s' % __version__,
                                       'errors': ['test error'],
                                       'filepath': 'test.py',
                                       'ast': None,
                                       'status': 'fatal'})


if __name__ == '__main__':
    unittest.main()
