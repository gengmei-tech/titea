#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2018 yongman <yming0221@gmail.com>
#
# Distributed under terms of the MIT license.

"""
unit test for set type
"""

import unittest
import time
import string
import random
from rediswrap import RedisWrapper

class SetTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.r = RedisWrapper().get_instance()
        cls.k1 = '__set1__'
        cls.k2 = '__set2__'
        cls.k3 = '__set3__'
        cls.v1 = 'value1'
        cls.v2 = 'value2'

    def setUp(self):
        self.r.delete(self.k1)
        self.r.delete(self.k2)
        self.r.delete(self.k3)
        pass

    def random_string(n):
        return ''.join(random.choice(string.ascii_uppercase + string.digits) for _ in range(n))

    def test_sadd(self):
        for i in range(200):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        self.assertEqual(self.r.scard(self.k1), 200)
        # for i in range(200):
        #     self.assertEqual(self.r.sadd(self.k1, str(i)), 0)
        # self.assertEqual(self.r.scard(self.k1), 200)

        # type exists
        self.assertTrue(self.r.exists(self.k1), "saddCommand exists error1")
        self.assertEqual(self.r.type(self.k1), "set", "saddCommand type error1")

    def test_scard(self):
        self.assertEqual(self.r.scard(self.k1), 0)
        for i in range(200):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        self.assertEqual(self.r.scard(self.k1), 200)
        # for i in range(200):
        #     self.assertEqual(self.r.sadd(self.k1, str(i)), 0)
        # self.assertEqual(self.r.scard(self.k1), 200)

    def test_sismember(self):
        for i in range(100):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        for i in range(100):
            self.assertEqual(self.r.sismember(self.k1, str(i)), 1)
        for i in range(100, 200):
            self.assertEqual(self.r.sismember(self.k1, str(i)), 0)

    def test_smembers(self):
        for i in range(200):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        self.assertSetEqual(self.r.smembers(self.k1), set([str(i) for i in range(200)]))

    def test_srem(self):
        for i in range(200):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        for i in range(10,100):
            self.assertEqual(self.r.srem(self.k1, str(i)), 1)
            self.assertEqual(self.r.scard(self.k1), 199+10-i)

    def test_sdiff(self):
        for i in range(0, 150):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        for i in range(100, 250):
            self.assertEqual(self.r.sadd(self.k2, str(i)), 1)
        self.assertSetEqual(self.r.sdiff(self.k1, self.k2), set([str(i) for i in range(0, 100)]))

    def test_sunion(self):
        for i in range(0, 150):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        for i in range(100, 250):
            self.assertEqual(self.r.sadd(self.k2, str(i)), 1)
        self.assertSetEqual(self.r.sunion(self.k1, self.k2), set([str(i) for i in range(0, 250)]))

    def test_sinter(self):
        for i in range(0, 150):
            self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
        for i in range(100, 250):
            self.assertEqual(self.r.sadd(self.k2, str(i)), 1)
        self.assertSetEqual(self.r.sinter(self.k1, self.k2), set([str(i) for i in range(100, 150)]))

    # def test_sdiffstore(self):
    #     for i in range(0, 150):
    #         self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
    #     for i in range(100, 250):
    #         self.assertEqual(self.r.sadd(self.k2, str(i)), 1)
    #     self.assertEqual(self.r.sdiffstore(self.k3, self.k1, self.k2), 100)
    #     self.assertSetEqual(self.r.smembers(self.k3), set([str(i) for i in range(0, 100)]))

    #     # type exists
    #     self.assertTrue(self.r.exists(self.k3), "sdiffstoreCommand exists error1")
    #     self.assertEqual(self.r.type(self.k3), "set", "sdiffstoreCommand type error1")

    # def test_sunionstore(self):
    #     for i in range(0, 150):
    #         self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
    #     for i in range(100, 250):
    #         self.assertEqual(self.r.sadd(self.k2, str(i)), 1)
    #     self.assertEqual(self.r.sunionstore(self.k3, self.k1, self.k2), 250)
    #     self.assertSetEqual(self.r.smembers(self.k3), set([str(i) for i in range(0, 250)]))

    #     # type exists
    #     self.assertTrue(self.r.exists(self.k3), "sunionstoreCommand exists error1")
    #     self.assertEqual(self.r.type(self.k3), "set", "sunionstoreCommand type error1")


    # def test_sinterstore(self):
    #     for i in range(0, 150):
    #         self.assertEqual(self.r.sadd(self.k1, str(i)), 1)
    #     for i in range(100, 250):
    #         self.assertEqual(self.r.sadd(self.k2, str(i)), 1)
    #     self.assertEqual(self.r.sinterstore(self.k3, self.k1, self.k2), 50)
    #     self.assertSetEqual(self.r.smembers(self.k3), set([str(i) for i in range(100, 150)]))

    #     # type exists
    #     self.assertTrue(self.r.exists(self.k3), "sinterstoreCommand exists error1")
    #     self.assertEqual(self.r.type(self.k3), "set", "sinterstoreCommand type error1")


    def test_del(self):
        self.assertEqual(self.r.sadd(self.k1, self.v1), 1)
        self.assertEqual(self.r.sadd(self.k2, self.v2), 1)

        self.assertEqual(self.r.delete(self.k1, self.k2), 2)

        self.assertEqual(self.r.scard(self.k1), 0)
        self.assertEqual(self.r.scard(self.k2), 0)

        # type exists
        self.assertFalse(self.r.exists(self.k1), "delCommand exists error1")
        self.assertFalse(self.r.exists(self.k2), "delCommand exists error2")

        self.assertIsNone(self.r.type(self.k1), "delCommand type error1")
        self.assertIsNone(self.r.type(self.k2), "delCommand type error1")


    def test_pexpire(self):
        self.assertEqual(self.r.sadd(self.k1, self.v1), 1)
        # expire in 5s
        self.assertTrue(self.r.pexpire(self.k1, 5000))
        self.assertLessEqual(self.r.pttl(self.k1), 5000)
        self.assertEqual(self.r.scard(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.scard(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "pexpireCommand set exists error1")
        self.assertIsNone(self.r.type(self.k1), "pexpireCommand set type error2")


    def test_pexpireat(self):
        self.assertEqual(self.r.sadd(self.k1, self.v1), 1)
        # expire in 5s
        ts = int(round(time.time()*1000)) + 5000
        self.assertTrue(self.r.pexpireat(self.k1, ts))
        self.assertLessEqual(self.r.pttl(self.k1), 5000)
        self.assertEqual(self.r.scard(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.scard(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "pexpireatCommand set exists error1")
        self.assertIsNone(self.r.type(self.k1), "pexpireatCommand set type error2")

    def test_expire(self):
        self.assertEqual(self.r.sadd(self.k1, self.v1), 1)
        # expire in 5s
        self.assertTrue(self.r.expire(self.k1, 5))
        self.assertLessEqual(self.r.ttl(self.k1), 5)
        self.assertEqual(self.r.scard(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.scard(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "expireCommand set exists error1")
        self.assertIsNone(self.r.type(self.k1), "expireCommand set type error2")


    def test_expireat(self):
        self.assertEqual(self.r.sadd(self.k1, self.v1), 1)
        # expire in 5s
        ts = int(round(time.time())) + 5
        self.assertTrue(self.r.expireat(self.k1, ts))
        self.assertLessEqual(self.r.ttl(self.k1), 5)
        self.assertEqual(self.r.scard(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.scard(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "expireatCommand set exists error1")
        self.assertIsNone(self.r.type(self.k1), "expireatCommand set type error2")

    def tearDown(self):
        pass

    @classmethod
    def tearDownClass(cls):
        cls.r.delete(cls.k1)
        cls.r.delete(cls.k2)
        cls.r.delete(cls.k3)
        print("clean up")
