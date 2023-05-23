# function that runs forever    

import time


def test():
    while True:
        print("Hello World2", flush=True)
        time.sleep(1)


test()