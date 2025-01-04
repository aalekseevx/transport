// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !wasm
// +build !wasm

package vnet

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pion/transport/v3/vtime"
	"github.com/pion/transport/v3/xtime"

	"github.com/pion/logging"
	"github.com/stretchr/testify/assert"
)

func TestTokenBucketFilter(t *testing.T) {
	t.Run("bitrateBelowCapacity", func(t *testing.T) {
		mnic := newMockNIC(t)

		tbf, err := NewTokenBucketFilter(mnic, TBFRate(10*MBit), TBFMaxBurst(10*MBit))
		assert.NoError(t, err, "should succeed")

		received := 0
		mnic.mockOnInboundChunk = func(Chunk) {
			received++
		}

		time.Sleep(1 * time.Second)

		sent := 100
		for i := 0; i < sent; i++ {
			tbf.onInboundChunk(&chunkUDP{
				userData: make([]byte, 1200),
			})
		}

		assert.NoError(t, tbf.Close())

		assert.Equal(t, sent, received)
	})

	subTest := func(t *testing.T, capacity int, maxBurst int, duration time.Duration, tm xtime.TimeManager) {
		log := logging.NewDefaultLoggerFactory().NewLogger("test")

		mnic := newMockNIC(t)

		tbf, err := NewTokenBucketFilter(mnic, TBFRate(capacity), TBFMaxBurst(maxBurst), TBFTimeManager(tm))
		assert.NoError(t, err, "should succeed")

		chunkChan := make(chan Chunk)
		mnic.mockOnInboundChunk = func(c Chunk) {
			chunkChan <- c
		}

		var wg sync.WaitGroup
		wg.Add(1)

		ctx, cancel := context.WithCancel(context.Background())

		rateCheckTicker := tm.NewTicker(time.Second)
		defer rateCheckTicker.Stop()

		go func() {
			defer wg.Done()

			totalBytesReceived := 0
			totalPacketsReceived := 0
			bytesReceived := 0
			packetsReceived := 0
			start := tm.Now()
			last := tm.Now()

			for {
				select {
				case <-ctx.Done():
					bits := float64(totalBytesReceived) * 8.0
					rate := bits / tm.Since(start).Seconds()
					mBitPerSecond := rate / float64(MBit)
					// Allow 5% more than capacity due to max bursts
					assert.Less(t, rate, 1.05*float64(capacity))
					assert.Greater(t, rate, 0.9*float64(capacity))

					log.Infof("duration=%v, bytesReceived=%v, packetsReceived=%v throughput=%.2f Mb/s", tm.Since(start), bytesReceived, packetsReceived, mBitPerSecond)
					return

				case now := <-rateCheckTicker.C():
					delta := now.Time().Sub(last)
					last = now.Time()
					bits := float64(bytesReceived) * 8.0
					rate := bits / delta.Seconds()
					mBitPerSecond := rate / float64(MBit)
					log.Infof("duration=%v, bytesReceived=%v, packetsReceived=%v throughput=%.2f Mb/s", delta, bytesReceived, packetsReceived, mBitPerSecond)
					// Allow 10% more than capacity due to max bursts
					fmt.Println("rate", rate, "capacity", capacity)
					assert.Less(t, rate, 1.10*float64(capacity))
					assert.Greater(t, rate, 0.9*float64(capacity))
					bytesReceived = 0
					packetsReceived = 0
					now.Done()

				case c := <-chunkChan:
					bytesReceived += len(c.UserData())
					packetsReceived++
					totalBytesReceived += len(c.UserData())
					totalPacketsReceived++
				}
			}
		}()

		senderTicker := tm.NewTicker(time.Millisecond)
		defer senderTicker.Stop()

		go func() {
			defer cancel()
			bytesSent := 0
			packetsSent := 0

			start := tm.Now()
			for tick := range senderTicker.C() {
				if tick.Time().Sub(start) > duration {
					return
				}
				c := &chunkUDP{
					userData: make([]byte, 1200),
				}
				tbf.onInboundChunk(c)
				bytesSent += len(c.UserData())
				packetsSent++
				tick.Done()
			}
			bits := float64(bytesSent) * 8.0
			rate := bits / tm.Since(start).Seconds()
			mBitPerSecond := rate / float64(MBit)
			log.Infof("duration=%v, bytesSent=%v, packetsSent=%v throughput=%.2f Mb/s", tm.Since(start), bytesSent, packetsSent, mBitPerSecond)

			assert.NoError(t, tbf.Close())
		}()

		if s, ok := tm.(*vtime.Simulator); ok {
			s.Start()
			defer s.Stop()
		}

		wg.Wait()
	}

	t.Run("500Kbit-s_StdTime", func(t *testing.T) {
		subTest(t, 500*KBit, 10*KBit, 10*time.Second, xtime.StdTimeManager{})
	})

	t.Run("1Mbit-s_StdTime", func(t *testing.T) {
		subTest(t, 1*MBit, 20*KBit, 10*time.Second, xtime.StdTimeManager{})
	})

	t.Run("2Mbit-s_StdTime", func(t *testing.T) {
		subTest(t, 2*MBit, 40*KBit, 10*time.Second, xtime.StdTimeManager{})
	})

	t.Run("8Mbit-s_StdTime", func(t *testing.T) {
		subTest(t, 8*MBit, 160*KBit, 10*time.Second, xtime.StdTimeManager{})
	})

	t.Run("500Kbit-s_VTime", func(t *testing.T) {
		tm := vtime.NewSimulator(time.Time{})
		subTest(t, 500*KBit, 10*KBit, 10*time.Second, tm)
	})

	t.Run("1Mbit-s_VTime", func(t *testing.T) {
		tm := vtime.NewSimulator(time.Time{})
		subTest(t, 1*MBit, 20*KBit, 10*time.Second, tm)
	})

	t.Run("2Mbit-s_VTime", func(t *testing.T) {
		tm := vtime.NewSimulator(time.Time{})
		subTest(t, 2*MBit, 40*KBit, 10*time.Second, tm)
	})

	t.Run("8Mbit-s_VTime", func(t *testing.T) {
		tm := vtime.NewSimulator(time.Time{})
		subTest(t, 8*MBit, 160*KBit, 10*time.Second, tm)
	})
}
