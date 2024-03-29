# monredis
a go daemon that syncs mongodb to redis in realtime


实时同步mongo 数据到redis。理论上支持redis 大部分命令，SADD/SET/RPUSH 等。

monredis 是根据 monstache 改编而来，大多数配置参考monstache。

### start
> monredis -f config.toml --enable-http-server

### example config.toml
```toml
mongo-url = "mongodb://user:password@mongo-urls/admin?replicaSet=mgset-17327671"
redis-url = "redis://:password@redis-urls/2"
direct-read-namespaces = ["your-database.collection", "your-database1.collection2"]
change-stream-namespaces = ["your-database.collection", "your-database1.collection2"]
stats = true
dropped-collections = false
dropped-databases = false
resume = true
resume-write-unsafe = true
resume-name = "default"
resume-strategy = 0
index-as-update = true
cluster-name = '<your-cluster-name>'
exit-after-direct-reads = false

[[mapping]]
namespace = "your-database.collection"
index = "index-name"
command = "SET"
key = "key:{{.key1}}:key2:{{.key2}}:key3:{{.key3}}"
val = "{{ toString .id }}"
expire = -1 # default -1

[[script]]
namespace = "your-database.collection"
script = """
  module.exports = function(doc) {
  doc.id = !!doc.id ? doc.id: doc._id
  return doc;
}
"""

[[mapping]]
namespace = "your-database1.collection1"
index = "index-name"
command = "SADD"
key = "{{.key1}}-{{.key2}}"
val = "{{ toJson . }}"

[[script]]
namespace = "your-database1.collection1"
script = """
  module.exports = function(doc) {
  doc.id = !!doc.id ? doc.id: doc._id
  return doc;
}
"""
```

### options
   mapping/key|value  支持 `toLower`,`toString`, `toSet`, `toJson` 等方法。
   script 支持javascript 语法
   
 
参考连接：[monstache](https://github.com/rwynn/monstache)
