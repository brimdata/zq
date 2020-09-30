package ingest

import "github.com/brimsec/zq/zio/ndjsonio"

var suricataTC *ndjsonio.TypeConfig = &ndjsonio.TypeConfig{Descriptors: map[string][]interface{}{"alert_log": []interface{}{map[string]interface{}{"name": "alert", "type": []interface{}{map[string]interface{}{"name": "action", "type": "bstring"}, map[string]interface{}{"name": "category", "type": "bstring"}, map[string]interface{}{"name": "gid", "type": "uint64"}, map[string]interface{}{"name": "rev", "type": "uint64"}, map[string]interface{}{"name": "severity", "type": "uint16"}, map[string]interface{}{"name": "signature", "type": "bstring"}, map[string]interface{}{"name": "signature_id", "type": "uint64"}, map[string]interface{}{"name": "metadata", "type": []interface{}{map[string]interface{}{"name": "signature_severity", "type": "array[bstring]"}, map[string]interface{}{"name": "former_category", "type": "array[bstring]"}, map[string]interface{}{"name": "attack_target", "type": "array[bstring]"}, map[string]interface{}{"name": "deployment", "type": "array[bstring]"}, map[string]interface{}{"name": "affected_product", "type": "array[bstring]"}, map[string]interface{}{"name": "created_at", "type": "array[bstring]"}, map[string]interface{}{"name": "performance_impact", "type": "array[bstring]"}, map[string]interface{}{"name": "updated_at", "type": "array[bstring]"}, map[string]interface{}{"name": "malware_family", "type": "array[bstring]"}, map[string]interface{}{"name": "tag", "type": "array[bstring]"}}}}}, map[string]interface{}{"name": "app_proto", "type": "bstring"}, map[string]interface{}{"name": "dest_ip", "type": "ip"}, map[string]interface{}{"name": "dest_port", "type": "port"}, map[string]interface{}{"name": "src_ip", "type": "ip"}, map[string]interface{}{"name": "src_port", "type": "port"}, map[string]interface{}{"name": "event_type", "type": "bstring"}, map[string]interface{}{"name": "flow_id", "type": "uint64"}, map[string]interface{}{"name": "pcap_cnt", "type": "uint64"}, map[string]interface{}{"name": "proto", "type": "bstring"}, map[string]interface{}{"name": "timestamp", "type": "time"}, map[string]interface{}{"name": "tx_id", "type": "uint64"}, map[string]interface{}{"name": "icmp_code", "type": "uint64"}, map[string]interface{}{"name": "icmp_type", "type": "uint64"}, map[string]interface{}{"name": "community_id", "type": "bstring"}}}, Rules: []ndjsonio.Rule{ndjsonio.Rule{Name: "event_type", Value: "alert", Descriptor: "alert_log"}}}
