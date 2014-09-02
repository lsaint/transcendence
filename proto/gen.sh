mv logic.proto logic.proto.bak
python filter.py
protoc --go_out=. server.proto
protoc --go_out=. logic.proto
protoc --python_out=. logic.proto 
protoc --python_out=. server.proto 
mv logic.proto.bak logic.proto
