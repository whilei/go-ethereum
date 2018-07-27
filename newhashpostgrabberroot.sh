example_hash=$(cat tests/files/GeneralStateTests/stSystemOperationsTest/CreateHashCollision.json | json CreateHashCollision.post.Homestead[0].hash)

echo "$example_hash"
ag "$example_hash" tests/files/StateTests


