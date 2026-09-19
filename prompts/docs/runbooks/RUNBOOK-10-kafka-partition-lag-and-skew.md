# Runbook 10: Kafka Partition Skew & Consumer Lag Remediation

**Runbook ID:** RBK-OPS-010  
**Severity Tier:** P2 (High)  
**Authority:** Event Platform Engineer / SRE On-Call  

---

## 1. Description & Alert Triggers
- PromQL Alert: `KafkaPartitionLag{topic="growww.order.events.v1"} > 10000` on a single partition for $> 60\text{ seconds}$.
- Partition processing latency exceeding $50\text{ms}$.

---

## 2. Remediation Workflow
1. **Identify Hot Partition:**
   ```bash
   kafka-consumer-groups --bootstrap-server 127.0.0.1:9092 --describe --group matching-engine-group
   ```
2. **Enable Dynamic Sub-Partition Salting:**
   If a single bellwether instrument is overloading the partition, execute runtime admin command:
   ```bash
   curl -X POST https://127.0.0.1:8019/api/v1/admin/kafka/salt-partition \
     -H "Authorization: Bearer ${OPERATOR_TOKEN}" \
     -d '{"instrument_id": "INE002A01018", "sub_partitions": 8}'
   ```
3. **Scale Consumer Workers:**
   Scale `matching-engine` consumer replicas to 32 instances to match partition concurrency.
