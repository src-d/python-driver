# To test with a high performance client sending the MSGPACK
# messages

import os
import strtabs
import msgpack4nim

when isMainModule:
  var msg = {"action": "ParseAST", "language": "python"}.newStringTable
  var f: File
  var files = commandLineParams()

  for filepath in files:
    discard f.open(filepath, fmRead)
    msg["filepath"] = filepath
    msg["content"] = f.readAll
    msg["
    var msgbuf = pack(msg)
    stdout.write(msgbuf)
