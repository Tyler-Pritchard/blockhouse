# blockhouse

## Redpanda Local Setup Instructions

### 1. Pull the Redpanda Image:
```
docker pull docker.vectorized.io/vectorized/redpanda:latest
```
### 2. Run Redpanda Container:
```
docker run -d --name redpanda-node \
  -p 9092:9092 \
  -p 9644:9644 \
  docker.vectorized.io/vectorized/redpanda:latest \
  redpanda start --overprovisioned --smp 1 --memory 1G \
  --reserve-memory 0M --node-id 0 --check=false \
  --kafka-addr PLAINTEXT://0.0.0.0:9092 \
  --advertise-kafka-addr PLAINTEXT://localhost:9092
```
### 3. Verify Cluster Health:
```
docker exec -it redpanda-node rpk cluster health
```
### 4. Create a Topic with Partitions:

- Replace test_topic with your desired topic name and adjust the partition count as needed.
```
docker exec -it redpanda-node rpk topic create test_topic --partitions 3
```
### 5. Produce and Consume Messages:

- Produce messages:
```
docker exec -it redpanda-node rpk topic produce test_topic
```
- Consume messages:
```
docker exec -it redpanda-node rpk topic consume test_topic
```