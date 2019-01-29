#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2018 yongman <yming0221@gmail.com>
#
# Distributed under terms of the MIT license.

"""
unit test for list type
"""

import unittest
import time
import string
import random
from rediswrap import RedisWrapper

class ListTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.r = RedisWrapper().get_instance()
        cls.k1 = '__list1__'
        cls.k2 = '__list2__'
        cls.v1 = 'value1'
        cls.v2 = 'value2'

    def setUp(self):
        self.r.delete(self.k1)
        self.r.delete(self.k2)
        pass

    def random_string(n):
        return ''.join(random.choice(string.ascii_uppercase + string.digits) for _ in range(n))

    def test_lpop(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
        for i in range(20):
            self.assertEqual(self.r.lpop(self.k1), str(i))

        # type exists
        self.assertFalse(self.r.exists(self.k1), "lpopCommand exists error1")
        self.assertIsNone(self.r.type(self.k1), "lpopCommand type error1")

    def test_lpush(self):
        for i in range(20):
            self.assertTrue(self.r.lpush(self.k1, str(i)))
        for i in range(20):
            self.assertEqual(self.r.rpop(self.k1), str(i))

        # type exists
        self.assertFalse(self.r.exists(self.k1), "lpushCommand exists error1")
        self.assertIsNone(self.r.type(self.k1), "lpushCommand type error1")

    def test_rpop(self):
        for i in range(20):
            self.assertTrue(self.r.lpush(self.k1, str(i)))
        for i in range(20):
            self.assertEqual(self.r.rpop(self.k1), str(i))

    def test_rpush(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
        for i in range(20):
            self.assertEqual(self.r.lpop(self.k1), str(i))

    def test_llen(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
        self.assertEqual(self.r.llen(self.k1), 20)

    def test_lindex(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
        for i in range(20):
            self.assertEqual(self.r.lindex(self.k1, i), str(i))

    def test_lrange(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
        self.assertListEqual(self.r.lrange(self.k1, 10, 19), [str(i) for i in range(10, 20)])

    def test_lset(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
        self.assertTrue(self.r.lset(self.k1, 10, 'hello'))
        self.assertEqual(self.r.lindex(self.k1, 10), 'hello')

    # def test_ltrim(self):
    #     for i in range(20):
    #         self.assertTrue(self.r.rpush(self.k1, str(i)))
    #     self.assertTrue(self.r.ltrim(self.k1, 0, 10))
    #     self.assertListEqual(self.r.lrange(self.k1, 0, -1), [str(i) for i in range(0, 11)])
    #     self.assertEqual(self.r.llen(self.k1), 11)

    def test_del(self):
        for i in range(20):
            self.assertTrue(self.r.rpush(self.k1, str(i)))
            self.assertTrue(self.r.rpush(self.k2, str(i)))

        self.assertEqual(self.r.delete(self.k1, self.k2), 2)


        self.assertFalse(self.r.exists(self.k1), "delCommand list exists error1")
        self.assertFalse(self.r.exists(self.k2), "delCommand list exists error2")

        self.assertIsNone(self.r.type(self.k1), "delCommand list type error1")
        self.assertIsNone(self.r.type(self.k2), "delCommand list type error2")


    def test_pexpire(self):
        self.assertTrue(self.r.lpush(self.k1, self.v1))
        # expire in 5s
        self.assertTrue(self.r.pexpire(self.k1, 5000))
        self.assertLessEqual(self.r.pttl(self.k1), 5000)
        self.assertEqual(self.r.llen(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.llen(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "pexpireCommand list exists error1")
        self.assertIsNone(self.r.type(self.k1), "pexpireCommand list type error2")

    def test_pexpireat(self):
        self.assertTrue(self.r.lpush(self.k1, self.v1))
        # expire in 5s
        ts = int(round(time.time()*1000)) + 5000
        self.assertTrue(self.r.pexpireat(self.k1, ts))
        self.assertLessEqual(self.r.pttl(self.k1), 5000)
        self.assertEqual(self.r.llen(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.llen(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "pexpireatCommand list exists error1")
        self.assertIsNone(self.r.type(self.k1), "pexpireatCommand list type error2")

    def test_expire(self):
        self.assertTrue(self.r.lpush(self.k1, self.v1))
        # expire in 5s
        self.assertTrue(self.r.expire(self.k1, 5))
        self.assertLessEqual(self.r.ttl(self.k1), 5)
        self.assertEqual(self.r.llen(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.llen(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "expireCommand list exists error1")
        self.assertIsNone(self.r.type(self.k1), "expireCommand list type error2")


    def test_expireat(self):
        self.assertTrue(self.r.lpush(self.k1, self.v1))
        # expire in 5s
        ts = int(round(time.time())) + 5
        self.assertTrue(self.r.expireat(self.k1, ts))
        self.assertLessEqual(self.r.ttl(self.k1), 5)
        self.assertEqual(self.r.llen(self.k1), 1)
        time.sleep(6)
        self.assertEqual(self.r.llen(self.k1), 0)

        self.assertFalse(self.r.exists(self.k1), "expireCommand list exists error1")
        self.assertIsNone(self.r.type(self.k1), "expireCommand list type error2")

    def tearDown(self):
        pass

    @classmethod
    def tearDownClass(cls):
        cls.r.delete(cls.k1)
        cls.r.delete(cls.k2)
        print("clean up")
