# -*- coding: utf-8 -*-

import salcb
from logic_pb2 import *

def simulateSalClientProto():
    print "test-simulateSalClientProto"
    pb = C2SLogin()
    pb.uid = 900
    salcb.OnSALClientProto(1640285, 50001906, 101, pb.SerializeToString())
