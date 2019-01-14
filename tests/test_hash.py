#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2018 yongman <yming0221@gmail.com>
#
# Distributed under terms of the MIT license.

"""
unit test for hash type
"""

import unittest
import time
from rediswrap import RedisWrapper

class HashTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.r = RedisWrapper().get_instance()
        cls.k1 = '__hash1__'
        cls.k2 = '__hash2__'

        cls.f1 = 'f1'
        cls.f2 = 'f2'
        cls.f3 = 'f3'
        cls.f4 = 'f4'

        cls.v1 = 'value1'
        cls.v2 = 'value2'
        cls.v3 = 'value3'
        cls.v4 = 'value4'

    def setUp(self):
        self.r.delete(self.k1)
        self.r.delete(self.k2)
        pass

    def test_hget(self):
        self.assertEqual(self.r.hset(self.k1, self.f1, self.v1), 1)
        self.assertEqual(self.v1, self.r.hget(self.k1, self.f1))



    def test_hset(self):
        self.assertEqual(self.r.hset(self.k1, self.f1, self.v1), 1)
        self.assertEqual(self.v1, self.r.hget(self.k1, self.f1))

        #type exists
        self.assertTrue(self.r.exists(self.k1), "hsetCommand exists error1")
        self.assertEqual(self.r.type(self.k1), "hash", "hsetCommand type error1")


    def test_hexists(self):
        self.assertEqual(self.r.hset(self.k1, self.f1, self.v1), 1)
        self.assertTrue(self.r.hexists(self.k1, self.f1))

    def test_hstrlen(self):
        self.assertEqual(self.r.hset(self.k1, self.f1, self.v1), 1)
        self.assertEqual(self.r.hstrlen(self.k1, self.f1), len(self.v1))

    def test_hlen(self):
        prefix = '__'
        for i in range(0, 20):
            f = '{}{}'.format(prefix, i)
            self.assertEqual(self.r.hset(self.k2, f, f), 1)
        self.assertEqual(self.r.hlen(self.k2), 20)

    def test_hmget(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))
        self.assertListEqual(self.r.hmget(self.k1, self.f1, self.f2, self.f3), [self.v1, self.v2, self.v3])

    def test_hkeys(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))
        self.assertListEqual(self.r.hkeys(self.k1), [self.f1, self.f2, self.f3])

        self.assertTrue(self.r.exists(self.k1), "hkeysCommand exists error1")
        self.assertEqual(self.r.type(self.k1), "hash", "hkeysCommand hash type error1")

    def test_hvals(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))
        self.assertListEqual(self.r.hvals(self.k1), [self.v1, self.v2, self.v3])

    def test_hgetall(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))
        self.assertDictEqual(self.r.hgetall(self.k1), {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3})


    def test_del(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1: self.v1, self.f2: self.v2, self.f3: self.v3}))
        self.assertTrue(self.r.hmset(self.k2, {self.f1: self.v1, self.f2: self.v2, self.f3: self.v3}))

        self.assertEqual(self.r.delete(self.k1, self.k2), 2, 'delCommand hash excute error')
        self.assertEqual(len(self.r.hgetall(self.k1)), 0, 'delCommand hash not delete1')
        self.assertEqual(len(self.r.hgetall(self.k2)), 0, 'delCommand hash not delete2')
        self.assertEqual(self.r.hlen(self.k1), 0)
        self.assertEqual(self.r.hlen(self.k2), 0)

        # 判断exists
        self.assertFalse(self.r.exists(self.k1), "delCommand hash exists error1")
        self.assertFalse(self.r.exists(self.k2), "delCommand hash exists error2")

        # 判断类型type
        self.assertIsNone(self.r.type(self.k1), "delCommand hash type error1")
        self.assertIsNone(self.r.type(self.k2), "delCommand hash type error2")

    def test_pexpire(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))
        # expire in 5s
        self.assertEqual(self.r.pexpire(self.k1, 5000), 1)
        self.assertLessEqual(self.r.pttl(self.k1), 5000)
        self.assertEqual(self.r.hlen(self.k1), 3)
        time.sleep(6)
        self.assertEqual(self.r.hlen(self.k1), 0)

        self.assertEqual(len(self.r.hgetall(self.k1)), 0, 'pexpireCommand hash getall error')

        self.assertFalse(self.r.exists(self.k1), "pexpireCommand hash exists error1")
        self.assertIsNone(self.r.type(self.k1), "pexpireCommand hash type error2")

    def test_pexpireat(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))

        # expire in 5s
        ts = int(round(time.time()*1000)) + 5000
        self.assertEqual(self.r.pexpireat(self.k1, ts), 1)
        self.assertLessEqual(self.r.pttl(self.k1), 5000)
        self.assertEqual(self.r.hlen(self.k1), 3)
        time.sleep(6)
        self.assertEqual(self.r.hlen(self.k1), 0)

        self.assertEqual(len(self.r.hgetall(self.k1)), 0, 'pexpireatCommand hash getall error')

        self.assertFalse(self.r.exists(self.k1), "pexpireatCommand hash exists error1")
        self.assertIsNone(self.r.type(self.k1), "pexpireatCommand hash type error2")


    def test_expire(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))

        # expire in 5s
        self.assertEqual(self.r.expire(self.k1, 5), 1)
        self.assertLessEqual(self.r.ttl(self.k1), 5)
        self.assertEqual(self.r.hlen(self.k1), 3)
        time.sleep(6)
        self.assertEqual(self.r.hlen(self.k1), 0)

        self.assertEqual(len(self.r.hgetall(self.k1)), 0, 'expireCommand hash getall error')

        self.assertFalse(self.r.exists(self.k1), "expireCommand hash exists error1")
        self.assertIsNone(self.r.type(self.k1), "expireCommand hash type error2")

    def test_expireat(self):
        self.assertTrue(self.r.hmset(self.k1, {self.f1:self.v1, self.f2:self.v2, self.f3:self.v3}))
        # expire in 5s
        ts = int(round(time.time())) + 5
        self.assertEqual(self.r.expireat(self.k1, ts), 1)
        self.assertLessEqual(self.r.ttl(self.k1), 5)
        self.assertEqual(self.r.hlen(self.k1), 3)
        time.sleep(6)
        self.assertEqual(self.r.hlen(self.k1), 0)

        self.assertDictEqual(self.r.hgetall(self.k1), {}, 'expireatCommand hash getall error')

        self.assertFalse(self.r.exists(self.k1), "expireatCommand hash exists error1")
        self.assertIsNone(self.r.type(self.k1), "expireatCommand hash type error2")


    def tearDown(self):
        pass

    @classmethod
    def tearDownClass(cls):
        cls.r.delete(cls.k1)
        cls.r.delete(cls.k2)
        print("clean up")
