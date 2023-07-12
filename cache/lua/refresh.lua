--1.检查是不是你的锁
--2.增加过期时间
-- KEYS[1] 就是你的分布式锁的key
-- ARGV[1] 就是你预期的存在redis里面的value

if redis.call('get', KEYS[1]) == ARGV[1] then 
    -- 确实是你的锁
    return redis.call('EXPIRE', KEYS[1], ARGV[2])
else
    -- 不是你的锁
    return 0
end