# function that runs forever    

import time


def test():
    wow = 0
    while True:
        if wow == 20:
            break
        print("Hello World " + str(wow), flush=True)
        time.sleep(1)
        wow = wow + 1


test()