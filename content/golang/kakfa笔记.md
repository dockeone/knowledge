1. /etc/hosts中需要绑定主机名和IP的对应关系，否则提示：
Getting: In the middle of a leadership election, there is currently no leader for this partition and hence it is unavailable for writes

生产者：
package main

import (
    "fmt"
    "github.com/Shopify/sarama"
)

func main() {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Partitioner = sarama.NewRandomPartitioner

    producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
    if err != nil {
        panic(err)
    }

    defer producer.Close()

    msg := &sarama.ProducerMessage {
        Topic: "kltao",
        Partition: int32(-1),
        Key:        sarama.StringEncoder("key"),
    }

    var value string
    for {
        _, err := fmt.Scanf("%s", &value)
        if err != nil {
            break
        }
        msg.Value = sarama.ByteEncoder(value)
        fmt.Println(value)

        partition, offset, err := producer.SendMessage(msg)
        if err != nil {
            fmt.Println("Send message Fail")
        }
        fmt.Printf("Partition = %d, offset=%d\n", partition, offset)
    }
}

消费者
package main

import (
    "fmt"
    "sync"

    "github.com/Shopify/sarama"
)

var (
    wg  sync.WaitGroup
)

func main() {
    consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, nil)
    if err != nil {
        panic(err)
    }

    partitionList, err := consumer.Partitions("kltao")
    if err != nil {
        panic(err)
    }

    for partition := range partitionList {
        pc, err := consumer.ConsumePartition("kltao", int32(partition), sarama.OffsetNewest)
        if err != nil {
            panic(err)
        }

        defer pc.AsyncClose()

        wg.Add(1)

        go func(sarama.PartitionConsumer) {
            defer wg.Done()
            for msg := range pc.Messages() {
                fmt.Printf("Partition:%d, Offset:%d, Key:%s, Value:%s\n", msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
            }

        }(pc)
    }
    wg.Wait()
    consumer.Close()
}