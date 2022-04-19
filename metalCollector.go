package main

import (
	"fmt"
	"strconv"
	"strings"

	metalgo "github.com/metal-stack/metal-go"

	"github.com/prometheus/client_golang/prometheus"
)

type metalCollector struct {
	networkInfo       *prometheus.Desc
	usedIps           *prometheus.Desc
	availableIps      *prometheus.Desc
	usedPrefixes      *prometheus.Desc
	availablePrefixes *prometheus.Desc

	partitionCapacityTotal     *prometheus.Desc
	partitionCapacityFree      *prometheus.Desc
	partitionCapacityAllocated *prometheus.Desc
	partitionCapacityFaulty    *prometheus.Desc

	driver *metalgo.Driver
}

func newMetalCollector(driver *metalgo.Driver) *metalCollector {
	return &metalCollector{
		networkInfo: prometheus.NewDesc("metal_network_info",
			"Shows available prefixes in a network",
			[]string{"networkId", "name", "projectId", "description", "partition", "vrf", "prefixes", "destPrefixes", "parentNetworkID", "isPrivateSuper", "useNat", "isUnderlay"}, nil,
		),
		usedIps: prometheus.NewDesc(
			"metal_network_ip_used",
			"The total number of used IPs of the network",
			[]string{"networkId"}, nil,
		),
		availableIps: prometheus.NewDesc(
			"metal_network_ip_available",
			"The total number of available IPs of the network",
			[]string{"networkId"}, nil,
		),
		usedPrefixes: prometheus.NewDesc(
			"metal_network_prefix_used",
			"The total number of used prefixes of the network",
			[]string{"networkId"}, nil,
		),
		availablePrefixes: prometheus.NewDesc(
			"metal_network_prefix_available",
			"The total number of available prefixes of the network",
			[]string{"networkId"}, nil,
		),
		partitionCapacityTotal: prometheus.NewDesc(
			"metal_partition_partitionCapacity_total",
			"The total partitionCapacity of machines in the partition",
			[]string{"partition", "size"}, nil,
		),
		partitionCapacityFree: prometheus.NewDesc(
			"metal_partition_partitionCapacity_free",
			"The partitionCapacity of free machines in the partition",
			[]string{"partition", "size"}, nil,
		),
		partitionCapacityAllocated: prometheus.NewDesc(
			"metal_partition_partitionCapacity_allocated",
			"The partitionCapacity of allocated machines in the partition",
			[]string{"partition", "size"}, nil,
		),
		partitionCapacityFaulty: prometheus.NewDesc(
			"metal_partition_partitionCapacity_faulty",
			"The partitionCapacity of faulty machines in the partition",
			[]string{"partition", "size"}, nil,
		),

		driver: driver,
	}
}

func (collector *metalCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.networkInfo
}

func (collector *metalCollector) Collect(ch chan<- prometheus.Metric) {
	networks, err := collector.driver.NetworkList()
	if err != nil {
		panic(err)
	}
	for _, n := range networks.Networks {
		privateSuper := fmt.Sprintf("%t", *n.Privatesuper)
		nat := fmt.Sprintf("%t", *n.Nat)
		underlay := fmt.Sprintf("%t", *n.Underlay)
		prefixes := strings.Join(n.Prefixes, ",")
		destPrefixes := strings.Join(n.Destinationprefixes, ",")
		vrf := strconv.FormatUint(uint64(n.Vrf), 10)

		// {"networkId", "name", "projectId", "description", "partition", "vrf", "prefixes", "destPrefixes", "parentNetworkID", "isPrivateSuper", "useNat", "isUnderlay"}, nil,
		ch <- prometheus.MustNewConstMetric(collector.networkInfo, prometheus.GaugeValue, 1.0, *n.ID, n.Name, n.Projectid, n.Description, n.Partitionid, vrf, prefixes, destPrefixes, n.Parentnetworkid, privateSuper, nat, underlay)
		ch <- prometheus.MustNewConstMetric(collector.usedIps, prometheus.GaugeValue, float64(*n.Usage.UsedIps), *n.ID)
		ch <- prometheus.MustNewConstMetric(collector.availableIps, prometheus.GaugeValue, float64(*n.Usage.AvailableIps), *n.ID)
		ch <- prometheus.MustNewConstMetric(collector.usedPrefixes, prometheus.GaugeValue, float64(*n.Usage.UsedPrefixes), *n.ID)
		ch <- prometheus.MustNewConstMetric(collector.availablePrefixes, prometheus.GaugeValue, float64(*n.Usage.AvailablePrefixes), *n.ID)
	}

	caps, err := collector.driver.PartitionCapacity(metalgo.PartitionCapacityRequest{})
	if err != nil {
		panic(err)
	}
	for _, p := range caps.Capacity {
		for _, s := range p.Servers {
			ch <- prometheus.MustNewConstMetric(collector.partitionCapacityTotal, prometheus.GaugeValue, float64(*s.Total), *p.ID, *s.Size)
			ch <- prometheus.MustNewConstMetric(collector.partitionCapacityAllocated, prometheus.GaugeValue, float64(*s.Allocated), *p.ID, *s.Size)
			ch <- prometheus.MustNewConstMetric(collector.partitionCapacityFree, prometheus.GaugeValue, float64(*s.Free), *p.ID, *s.Size)
			ch <- prometheus.MustNewConstMetric(collector.partitionCapacityFaulty, prometheus.GaugeValue, float64(*s.Faulty), *p.ID, *s.Size)
		}
	}
}
