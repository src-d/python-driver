#!/usr/bin/env python3

import sys
import signal
import logging
from traceback import print_exc
from python_driver.processor_configs import ProcessorConfigs
from python_driver.requestprocessor import RequestProcessor, InBuffer, OutBuffer
from typing import Any, Tuple, Optional

logging.basicConfig(filename="python_driver.log", level=logging.ERROR)


class RequestInstantiationException(Exception):
    pass


# Gracefully handle control c without adding another try-except on top of the loop
def ctrlc_signal_handler(sgn: int, frame: Any) -> None:
    sys.exit(0)
signal.signal(signal.SIGINT, ctrlc_signal_handler)


def get_processor_instance(format_: str, custom_inbuffer: InBuffer=None,
               custom_outbuffer: OutBuffer=None) -> Tuple[Any, Any]:
    """
    Get a processor instance. The class and buffers will be selected based on the
    python_driver.ProcessorConfigs dictionary. The input and output buffers can
    be overriden using the custom_inbuffer and custom_outbuffer parameters. This
    is mainly useful for unittesting.
    """
    conf = ProcessorConfigs.get(format_)
    if not conf:
        raise RequestInstantiationException('No RequestProcessor found for format %s' % format)

    inbuffer  = custom_inbuffer if custom_inbuffer else conf['inbuffer']
    outbuffer = custom_outbuffer if custom_outbuffer else conf['outbuffer']
    instance  = conf['class'](outbuffer) # type: ignore

    return instance, inbuffer


def main() -> None:
    """
    If you pass the --json command line parameter the replies will be
    printed using pprint without wrapping them in the msgpack format which
    can be handy when debugging.
    """

    if len(sys.argv) > 1 and sys.argv[1] == '--json':
        format_ = 'json'
    else:
        format_ = 'msgpack'

    processor, inbuffer = get_processor_instance(format_)
    try:
        processor.process_requests(inbuffer)
    except UnicodeDecodeError:
        print_exc()
        print('Error while trying to decode the message, are you sure you are not '
              'using a different input format that the currently configured (%s)?' % format)


if __name__ == '__main__':
    main()
