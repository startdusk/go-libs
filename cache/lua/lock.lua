local val = redis.call('get', KEYS[1])
if val == false then
-- key不存在, set返回OK
    return redis.call('set', KEYS[1], ARGV[1], 'EX', ARGV[2])
elseif val == ARGV[1] then
-- 这里代表你上次加锁成功了, expire返回数字1, 所以为了保持统一返回'OK'
    redis.call('expire', KEYS[1], ARGV[2])
    return 'OK'
else
-- 说明锁被人拿着
    return ''
end
