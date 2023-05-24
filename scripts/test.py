import time
import os
import shutil
import sys

taskId = sys.argv[1]
IN_PATH = "../rep/" + taskId + "/input/"
OUT_PATH = "../rep/" + taskId + "/output/"

def test():
    if not os.path.exists(IN_PATH):
        raise Exception("Input path does not exist")

    if not os.path.exists(OUT_PATH):
        os.makedirs(OUT_PATH)
    for file in os.listdir(IN_PATH):
        shutil.copy2(IN_PATH + file, OUT_PATH)

    i = 0
    while True:
        if i == 20:
            break
        print("Hello World " + str(i), flush=True)
        i = i + 1
        time.sleep(1)

test()
