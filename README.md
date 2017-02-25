This is the Python driver for bblfsh. It will try to detect the actual Python
version on a best-effort to include it in the reply, defaulting to Python3 for
codefiles that seem to be compatible with both (the AST is better). 

It follows the standard bblfsh driver protocol, receiving files form standard input
and sending replies on stdout supporting the msgpack or json formats.

The performance parsing Python standard library is about 90 msecs/file on average.
It could surely be improved by using less regular expressions to detect the Python
version, but then accuracy of the detection would suffer. Anyway there is surely
room for other optimizations once an extensive profiling has been done.
