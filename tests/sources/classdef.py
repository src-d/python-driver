class Animal:
    def __init__(self):
        self._a = 1
        b = 2

    def method(self, arg1):
        pass

    @property
    def a():
        return self._a

    @a.setter
    def a(newa):
        self._a = newa

class Dog(Animal):
    pass

a = Animal()
b = Dog()
a.a = 1
print(a.a)

