#!/bin/bash

# Инициализация бд

docker compose exec -T config_srv mongosh --port 27010 --quiet <<EOF
rs.initiate(
  {
    _id : "config",
    configsvr: true,
    members: [
      { _id : 0, host : "config_srv:27010" }
    ]
  }
);
EOF

docker compose exec -T shard_1-replica_1 mongosh --port 27011 --quiet <<EOF
rs.initiate(
    {
      _id : "shard_1",
      members: [
        { _id : 0, host : "shard_1-replica_1:27011" },
        { _id : 1, host : "shard_1-replica_2:27012" },
        { _id : 2, host : "shard_1-replica_3:27013" }

      ]
    }
);
EOF

docker compose exec -T shard_2-replica_1 mongosh --port 27021 --quiet <<EOF
rs.initiate(
    {
      _id : "shard_2",
      members: [
        { _id : 0, host : "shard_2-replica_1:27021" },
        { _id : 1, host : "shard_2-replica_2:27022" },
        { _id : 2, host : "shard_2-replica_3:27023" }
      ]
    }
  );
EOF

docker compose exec -T mongos_router mongosh --port 27020 --quiet <<EOF
sh.addShard( "shard_1/shard_1-replica_1:27011,shard_1-replica_2:27012,shard_1-replica_3:27013");
sh.addShard( "shard_2/shard_2-replica_1:27021,shard_2-replica_2:27022,shard_2-replica_3:27023");
sh.enableSharding("somedb");
sh.shardCollection("somedb.helloDoc", { "name" : "hashed" } )
EOF

# Инициализация данных

docker compose exec -T mongos_router mongosh --port 27020 --quiet <<EOF
use somedb;
for(var i = 0; i < 1111; i++) db.helloDoc.insert({age:i, name:"pipa_"+i})
EOF