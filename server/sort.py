# This Python file uses the following encoding: utf-8
import os, sys

file_name = open("./lat.out", "r")

br_lats = [[],[],[],[]]

for l in file_name:
    l_sp = l.split(' ')
    if len(l_sp) > 2:
#        print(l_sp[2])
#        print(l_sp[2][:-3])
#        print(l_sp[2][-3:])
        if l_sp[2][-3:] == "ns\n":
            br_lats[0].append(float(l_sp[2][:-3]))
        elif l_sp[2][-3:] == "µs\n":
             br_lats[1].append(float(l_sp[2][:-3]))
        elif l_sp[2][-3:] == "ms\n":
            br_lats[2].append(float(l_sp[2][:-3]) * 1000)
        elif l_sp[2][-2:] == "s\n":
            br_lats[3].append(float(l_sp[2][:-4]) * 1000 * 1000)

i = 0
for l in lats:
    l.sort()
    for val in l:
        print(str(i) + ", " + str(val))
        i+=1
