import io
import sys
import json
import msgpack
import unittest
from os.path import dirname, join

sys.path.append('../bin')
from python_driver import __version__, get_processor_instance
from python_driver.requestprocessor import (
        RequestProcessorMSGPack, RequestProcessorJSON, RequestCheckException
)


def convert_bytes(data, to_bytes=False):
    """
    Both in normal operation and with this tests data comes trough a bytestream so this is needed to
    recursively convert the msgpack incoming data (the pprint data is converted just decoding and using
    literal_eval)
    """
    if type(data) in (list, tuple):
        newlist = []
        for item in data:
            newlist.append(convert_bytes(item, to_bytes))
        return newlist
    elif isinstance(data, dict):
        newdict = {}
        for key, value in data.items():
            newvalue = convert_bytes(value, to_bytes)
            newkey = convert_bytes(key, to_bytes)
            newdict[newkey] = newvalue
        return newdict
    elif isinstance(data, bytes) and not to_bytes:
        return data.decode()
    elif isinstance(data, str) and to_bytes:
        return data.encode()
    return data


class TestPythonDriverBase(unittest.TestCase):
    def _restart_data(self, format_='json'):
        assert format_ in ('json', 'msgpack')

        with open(join(dirname(__file__), 'data', 'helloworld.py')) as f:
            testcode = f.read()

        self.data = {
            'filepath': 'test.py',
            'action': 'ParseAST',
            'content': testcode,
            'language': 'python',
        }

        bufferclass = io.StringIO if format_ == 'json' else io.BytesIO

        # This will mock the python_driver stdin
        self.sendbuffer = bufferclass()
        # This will mock the python_driver stdout
        self.recvbuffer = bufferclass()

    @staticmethod
    def _extract_docs(inbuffer):
        """
        This generator will read the inbuffer yielding the JSON
        docs when it finds the ending mark
        """
        for line in inbuffer.readlines():
            yield json.loads(line)

    def _loadResults(self, format_):
        """Read all msgpacks from the recvbuffer"""
        self.recvbuffer.seek(0)

        if format_ == 'json':
            res = [doc for doc in self._extract_docs(self.recvbuffer)]
        elif format_ == 'msgpack':
            res = [convert_bytes(msg) for msg in msgpack.Unpacker(self.recvbuffer)]
        else:
            assert False

        return res


class Test10ProcessRequestFunc(TestPythonDriverBase):

    def _add_to_buffer(self, count, format_):
        """Add count test msgpacks to the sendbuffer"""
        for i in range(count):
            msg = ''
            if format_ == 'msgpack':
                msg = msgpack.dumps(self.data)
            elif format_ == 'json':
                msg = json.dumps(self.data, ensure_ascii=False) + '\n'
            self.sendbuffer.write(msg)

        self.sendbuffer.flush()

    def _send_receive(self, nummsgs, outformat='msgpack', dataupdate=None,
                      restart_data=True):
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

    def _check_reply_dict(self, reply, has_errors=False):
        self.assertIsInstance(reply, dict)
        self.assertEqual(reply.get('driver'), 'python23:{}'.format(__version__))
        status = reply.get('status')

        if has_errors:
            assert status in ('ok', 'fatal')
            errors = reply.get('errors', list)
            self.assertIsInstance(errors, list)
            self.assertGreater(len(errors), 0)
        else:
            self.assertEqual(reply.get('language'), 'python')
            self.assertEqual(status, 'ok')
            self._check_AST_dict(reply)
            language_version = reply.get('language_version', -1)
            assert str(language_version) in ('2', '3')

    def _check_AST_dict(self, reply):
        self.assertIsNotNone(reply)
        assert 'ast' in reply
        self.assertIsInstance(reply['ast'], dict)

        for key in ('ast_type', 'body'):
            assert key in reply['ast']

        self.assertIsInstance(reply['ast']['body'], list)

        for item in reply['ast']['body']:
            for key in ('ast_type', 'lineno', 'col_offset'):
                assert key in item

    def test_010_normal_json(self):
        replies = self._send_receive(1, 'json')
        self.assertEqual(len(replies), 1)
        self._check_reply_dict(replies[0])

    def test_020_normal_msgpack(self):
        replies = self._send_receive(1, 'msgpack')
        self.assertEqual(len(replies), 1)
        self._check_reply_dict(replies[0])

    def test_030_normal_json_many(self):
        replies = self._send_receive(100, 'json')
        self.assertEqual(len(replies), 100)
        for reply in replies:
            self._check_reply_dict(reply)

    def test_040_normal_msgpack_many(self):
        replies = self._send_receive(100, 'msgpack')
        self.assertEqual(len(replies), 100)
        for reply in replies:
            self._check_reply_dict(reply)

    def test_050_error_print(self):
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

    def test_060_error_msgpack(self):
        wrongcode = 'wtf lol'

        replies = self._send_receive(1, 'msgpack', {'content': wrongcode})
        self.assertEqual(len(replies), 1)
        ast = replies[0].get('ast')
        self.assertIsNone(ast)
        self._check_reply_dict(replies[0], has_errors=True)

        # Check that it still alive
        self._restart_data()
        replies = self._send_receive(1, 'json')
        self.assertEqual(len(replies), 1)

    def test_070_broken_msgpack(self):
        self._restart_data('msgpack')
        brokendata = msgpack.dumps(self.data)[:-30]
        self.sendbuffer.write(brokendata)
        self.sendbuffer.flush()
        reply = self._send_receive(1, 'msgpack', restart_data=False)[0]
        self.assertEqual(reply['status'], 'fatal')
        self.assertEqual(len(reply['errors']), 1)

    def test_080_broken_json(self):
        self._restart_data('json')
        brokendata = json.dumps(self.data, ensure_ascii=False)[:-30]
        self.sendbuffer.write(brokendata)
        self.sendbuffer.flush()
        reply = self._send_receive(1, 'json', restart_data=False)[0]
        self.assertEqual(reply['status'], 'fatal')
        self.assertEqual(len(reply['errors']), 1)


class Test20ReqProcMethods(TestPythonDriverBase):
    def test_check_input(self):
        self._restart_data('msgpack')
        brequest = convert_bytes(self.data, to_bytes=True)
        processor = RequestProcessorMSGPack(self.recvbuffer)
        res = processor._check_input_request(brequest)
        self.assertEqual(res[1], 'test.py')

    def test_check_input_bad(self):
        self._restart_data('msgpack')
        del self.data['content']
        brequest = convert_bytes(self.data, to_bytes=True)
        processor = RequestProcessorMSGPack(self.recvbuffer)
        with self.assertRaises(RequestCheckException) as _:  # noqa: F841
            processor._check_input_request(brequest)

    def test_send_response_msgpack(self):
        self._restart_data('msgpack')
        processor = RequestProcessorMSGPack(self.recvbuffer)
        processor._send_response(self.data)
        res = self._loadResults('msgpack')
        self.assertEqual(len(res), 1)
        self.assertDictEqual(self.data, res[0])

    def test_send_response_json(self):
        self._restart_data('json')
        processor = RequestProcessorJSON(self.recvbuffer)
        processor._send_response(self.data)
        res = self._loadResults('json')
        self.assertEqual(len(res), 1)
        self.assertDictEqual(self.data, res[0])

    # process request already tested with TestPythonDriverBase

    def test_return_error(self):
        self._restart_data('json')
        processor = RequestProcessorJSON(self.recvbuffer)
        processor.errors = ['test error']
        processor._return_error('test.py', 'fatal')
        res = self._loadResults('json')
        self.assertEqual(len(res), 1)
        self.assertDictEqual(res[0] , {'driver': 'python23:{}'.format(__version__),
                                       'errors': ['test error'],
                                       'filepath': 'test.py',
                                       'status': 'fatal'})


if __name__ == '__main__':
    unittest.main()
