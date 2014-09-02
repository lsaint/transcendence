

f = open("./logic.proto.bak", "r")
fw = open("./logic.proto", "w")
for line in f:
    if line.startswith("//"):
        continue
    sp = line.split("//")
    if len(sp) > 1:
        buf = sp[0] + "\n"
    else:
        buf = line
    fw.write(buf)

