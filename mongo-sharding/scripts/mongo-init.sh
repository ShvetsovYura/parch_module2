#!/bin/bash

# Инициализация бд

docker compose exec -T config_srv mongosh --port 27017 --quiet <<EOF
rs.initiate(
  {
    _id : "config",
    configsvr: true,
    members: [
      { _id : 0, host : "config_srv:27017" }
    ]
  }
);
EOF

docker compose exec -T shard_1 mongosh --port 27018 --quiet <<EOF
rs.initiate(
    {
      _id : "shard_1",
      members: [
        { _id : 0, host : "shard_1:27018" }
      ]
    }
);
EOF

docker compose exec -T shard_2 mongosh --port 27019 --quiet <<EOF
rs.initiate(
    {
      _id : "shard_2",
      members: [
        { _id : 1, host : "shard_2:27019" }
      ]
    }
  );
EOF

docker compose exec -T mongos_router mongosh --port 27020 --quiet <<EOF
sh.addShard( "shard_1/shard_1:27018");
sh.addShard( "shard_2/shard_2:27019");
sh.enableSharding("somedb");
sh.shardCollection("somedb.helloDoc", { "name" : "hashed" } )
EOF

# Инициализация данных

docker compose exec -T mongos_router mongosh --port 27020 --quiet <<EOF
use somedb;
for(var i = 0; i < 1111; i++) db.helloDoc.insert({age:i, name:"pipa_"+i})
EOF