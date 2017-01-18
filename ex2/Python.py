# Python 3.3.3 and 2.7.6
# python helloworld_python.py

from threading import Thread

i = 0

def thread1_func():
    global i
    for j in range(1000000):
	i -= 1

def thread2_func():
    global i
    for j in range(1000000):
	i += 1

# Potentially useful thing:
#   In Python you "import" a global variable, instead of "export"ing it when you declare it
#   (This is probably an effort to make you feel bad about typing the word "global")
#global i


def main():
    global i

    thread1 = Thread(target = thread1_func, args = (),)
    thread1.start()

    thread2 = Thread(target = thread2_func, args = (),)
    thread2.start()
    
    thread1.join()
    thread2.join()

    print("num:", i)


main()
