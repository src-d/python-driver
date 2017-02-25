import sys
from python_driver.requestprocessor import RequestProcessorJSON, RequestProcessorMSGPack

ProcessorConfigs = {
        'json': {
            'class': RequestProcessorJSON,
            'inbuffer': sys.stdin,
            'outbuffer': sys.stdout
        },

        'msgpack': {
            'class': RequestProcessorMSGPack,
            'inbuffer': sys.stdin.buffer,
            'outbuffer': sys.stdout.buffer
        }
}
