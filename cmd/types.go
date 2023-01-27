package cmd

import (
	"time"
)

type EventPacket struct {
	block_height             string
	block_time               string
	event_type               string
	packet_timeout_height    string
	packet_timeout_timestamp string
	packet_sequence          string
	packet_src_port          string
	packet_src_channel       string
	packet_dst_port          string
	packet_dst_channel       string
	packet_channel_ordering  string
	packet_connection        string
}

type PacketData struct {
	block_height       uint
	block_time         time.Time
	event_type         string
	packet_sequence    string
	packet_src_channel string
	packet_dst_channel string
	amount             uint
	denom              string
	receiver           string
	sender             string
}

type TimeoutData struct {
	block_height       uint
	block_time         time.Time
	event_type         string
	packet_sequence    string
	packet_src_channel string
	packet_dst_channel string
	module             string
	refund_receiver    string
	refund_denom       string
	refund_amount      uint
	memo               string
}
