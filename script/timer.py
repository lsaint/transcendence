# -*- coding: utf-8 -*-
import time, redis, config, go, cPickle

g_redis = redis.StrictRedis(host='localhost', port=6379, db=0)
g_redis.ping()

class Timer(object):
    TimerDict = {}    # {tid:[start_time, interval, loop, fun, args]}
    Index = 0

    # 记录本次实例化后创建的timerid
    def __init__(self):
        self._TimerIdList = []


    # 释放本次实例化后的所有timer
    def ReleaseTimer(self):
        for idx in self._TimerIdList[:]:
            self.KillTimer(idx)
            self._TimerIdList.remove(idx)


    @classmethod
    def Update(cls):
        if len(Timer.TimerDict) == 0:
            return
        t = time.time()
        for tid, lt in Timer.TimerDict.items():
            if t - lt[0] >= lt[1]:
                lt[0] = t
                if not lt[2]:
                    cls.KillTimer(tid)
                apply(lt[3], lt[4])


    def setTimer(self, sec, is_loop, callbackfun, *args):
        Timer.Index += 1
        mark = time.time()
        Timer.TimerDict[Timer.Index] = [mark, sec, is_loop, callbackfun, args]
        self._TimerIdList.append(Timer.Index)

        return Timer.Index


    def SetTimer1(self, sec, cb, *args):
        return self.setTimer(sec, False, cb, *args)


    def SetTimer(self, sec, cb, *args):
        return self.setTimer(sec, True, cb, *args)


    def DoSetTimer(self, sec, cb, *args):
        cb(*args)
        return self.setTimer(sec, True, cb, *args)


    @classmethod
    def KillTimer(cls, tid):
        if Timer.TimerDict.has_key(tid):
            del Timer.TimerDict[tid]


g_timer = Timer()


### ### ### ### ### ### ### 


class LeaderTimer(object):
    TimerDict = {}    
    Index = 0
    Key = config.REDIS_LEADER_TIMER_KEY

    def __init__(self):
        self._TimerIdList = []


    def ReleaseTimer(self):
        for idx in self._TimerIdList[:]:
            self.KillLeaderTimer(idx)
            self._TimerIdList.remove(idx)


    @classmethod
    def Update(cls):
        if len(cls.TimerDict) == 0:
            return
        t = time.time()
        tids_modify = []
        for tid, lt in cls.TimerDict.items():
            if t - lt[0] >= lt[1]:
                lt[0] = t
                if not lt[2]:
                    cls.KillTimer(tid)
                else:
                    tids_modify.append(tid)
                apply(lt[3], lt[4])


        if len(tids_modify) == 0:
            return

        # update in redis
        recoder = {}
        for tid in tids_modify:
            pt = cPickle.dumps(cls.TimerDict[tid])
            recoder[tid] = pt
        g_redis.hmset(cls.Key, recoder)


    @classmethod
    def OnBecomeLeader(cls):
        cache = g_redis.hgetall(cls.Key)
        print "cache", cache
        for tid, pt in cache.iteritems():
            cls.TimerDict[int(tid)] = cPickle.loads(pt)


    @classmethod
    def OnHandoffLeader(cls):
        cls.TimerDict = {}


    def setTimer(self, sec, is_loop, callbackfun, *args):
        if not go.IsLeader():
            return

        self.Index += 1
        mark = time.time()
        self.TimerDict[self.Index] = [mark, sec, is_loop, callbackfun, args]
        self._TimerIdList.append(self.Index)

        # set in redis
        pt = cPickle.dumps(self.TimerDict[self.Index])
        g_redis.hset(self.Key, self.Index, pt)

        return self.Index


    def SetLeaderTimer(self, sec, cb, *args):
        tid = self.setTimer(sec, True, cb, *args)


    def SetLeaderTimer1(self, sec, cb, *args):
        return self.setTimer(sec, False, cb, *args)


    def DoSetLeaderTimer(self, sec, cb, *args):
        cb(*args)
        return self.setTimer(sec, True, cb, *args)


    @classmethod
    def KillLeaderTimer(cls, tid):
        if not go.IsLeader():
            return

        if cls.TimerDict.has_key(tid):
            del cls.TimerDict[tid]
        # del in redis
        g_redis.hdel(cls.Key, tid)



g_leadertimer = LeaderTimer()

