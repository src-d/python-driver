#!/usr/bin/env python3
"""
Script to help test and debug. It will create a babelfish message
with the code of the files passed as argument and output them
on stdout so you can pipe them to python_driver or the docker image like
this:

$ ./sendmsg.py [source_file.py]|python_driver.py

Or with the Docker image:

$ ./sendmsg.py [source_file.py]|docker run -it --rm python_driver:latest
"""
import sys
import json
import msgpack
import logging
sys.path.append('../bin')
# logging.basicConfig(filename="sendmsg.log", level=logging.DEBUG)

def main():
    filesidx = 1
    if len(sys.argv) > 1 and sys.argv[1] == '--msgpack':
        outformat = 'msgpack'
        filesidx += 1
        outbuffer = sys.stdout.buffer
    else:
        outformat = 'json'
        outbuffer = sys.stdout

    files = sys.argv[filesidx:]

    d = {
        'action': 'ParseAST',
        'filepath': '',
        'content': '',
        'language': 'python',
    }

    for f in files:
        content = ''
        logging.info(f)
        for encoding in ('utf_8', 'iso8859_15', 'iso8859_15', 'gb2313',
                         'cp1251', 'cp1252', 'cp1250', 'shift-jis', 'gbk', 'cp1256',
                         'iso8859-2', 'euc_jp', 'big5', 'cp874', 'euc_kr', 'iso8859_7',
                         'cp1255'):
            with open(f, encoding=encoding) as infile:
                try:
                    content = infile.read()
                    break
                except UnicodeDecodeError:
                    continue

        d.update({
            'filepath': f,
            'content': content,
        })

        if outformat == 'json':
            json.dump(d, sys.stdout, ensure_ascii=False)
            outbuffer.write('\n')
        else:
            msg = msgpack.packb(d)
            outbuffer.write(msg)
    outbuffer.close()


if __name__ == '__main__':
    main()
