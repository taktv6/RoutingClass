package main

import (
	"flag"
	"fmt"
)

type Circuit struct {
	A       string
	B       string
	Address string
}

type Topology struct {
	L3Devices []string
	L2Devices []string
	Circuits  []Circuit
}

var (
	mode = flag.String("mode", "create", "create|destroy")
)

func main() {
	flag.Parse()

	t := Topology{
		L3Devices: []string{
			"S1",
			"S2",
			"R1",
			"R2",
			"R3",
			"R4",
			"R5",
			"R6",
		},
		L2Devices: []string{
			"LANA",
			"LANB",
		},
		Circuits: []Circuit{
			{
				A:       "S1",
				B:       "LANA",
				Address: "192.168.100.100/24",
			},
			{
				A:       "R1",
				B:       "LANA",
				Address: "192.168.100.1/24",
			},
			{
				A:       "R2",
				B:       "LANA",
				Address: "192.168.100.2/24",
			},
			{
				A:       "R1",
				B:       "R2",
				Address: "192.168.1.",
			},
			{
				A:       "R1",
				B:       "R3",
				Address: "192.168.2.",
			},
			{
				A:       "R2",
				B:       "R4",
				Address: "192.168.3.",
			},
			{
				A:       "R1",
				B:       "R4",
				Address: "192.168.4.",
			},
			{
				A:       "R2",
				B:       "R3",
				Address: "192.168.5.",
			},

			{
				A:       "R3",
				B:       "R5",
				Address: "192.168.6.",
			},
			{
				A:       "R4",
				B:       "R6",
				Address: "192.168.7.",
			},
			{
				A:       "R3",
				B:       "R6",
				Address: "192.168.8.",
			},
			{
				A:       "R4",
				B:       "R5",
				Address: "192.168.9.",
			},

			{
				A:       "R5",
				B:       "R6",
				Address: "192.168.10.",
			},

			{
				A:       "S2",
				B:       "LANB",
				Address: "192.168.200.100/24",
			},
			{
				A:       "R5",
				B:       "LANB",
				Address: "192.168.200.1/24",
			},
			{
				A:       "R6",
				B:       "LANB",
				Address: "192.168.200.2/24",
			},
		},
	}

	switch *mode {
	case "create":
		create(t)
	}

}

func create(t Topology) {
	for _, n := range t.L2Devices {
		createNetNS(n)
		createBridge(n, "br0")
		interfaceUp(n, "br0")
	}

	for _, n := range t.L3Devices {
		createNetNS(n)
		enableIPForwarding(n)
		disableRPF(n)
	}

	for _, c := range t.Circuits {
		createAdjacency(c)
	}

}

func createNetNS(n string) {
	fmt.Printf("ip netns add %s\n", n)
}

func createBridge(ns string, name string) {
	fmt.Printf("ip netns exec %s ip link add name %s type bridge\n", ns, name)
}

func interfaceUp(ns string, name string) {
	fmt.Printf("ip netns exec %s ip link set up dev %s\n", ns, name)
}

func enableIPForwarding(ns string) {
	fmt.Printf("ip netns exec %s bash -c \"for i in /proc/sys/net/ipv4/conf/*; do echo 1 > $i/forwarding; done\"\n", ns)
}

func disableRPF(ns string) {
	fmt.Printf("ip netns exec %s bash -c \"for i in /proc/sys/net/ipv4/conf/*; do echo 0 > $i/rp_filter; done\"\n", ns)
}

func createAdjacency(c Circuit) {
	x := fmt.Sprintf("%s-%s", c.A, c.B)
	y := fmt.Sprintf("%s-%s", c.B, c.A)

	fmt.Printf("ip link add %s type veth peer name %s\n", x, y)

	fmt.Printf("ip link set %s netns %s\n", x, c.A)
	interfaceUp(c.A, x)

	fmt.Printf("ip link set %s netns %s\n", y, c.B)
	interfaceUp(c.B, y)

	if c.B != "LANA" && c.B != "LANB" {
		addrA := c.Address + "0/31"
		addrB := c.Address + "1/31"

		fmt.Printf("ip netns exec %s ip addr add %s dev %s\n", c.A, addrA, x)
		fmt.Printf("ip netns exec %s ip addr add %s dev %s\n", c.B, addrB, y)
	} else {
		fmt.Printf("ip netns exec %s ip link set %s master br0\n", c.B, y)
		fmt.Printf("ip netns exec %s ip addr add %s dev %s\n", c.A, c.Address, x)
	}
}
