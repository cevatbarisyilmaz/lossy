/*

Package lossy simulates bandwidth, latency and packet loss for net.PacketConn and net.Conn interfaces.

*/
package lossy

//IPv4MinHeaderOverhead is the minimum header overhead for IPv4 based connections.
const IPv4MinHeaderOverhead = 20

//IPv4MaxHeaderOverhead is the maximum header overhead for IPv4 based connections.
const IPv4MaxHeaderOverhead = 60

//IPv6HeaderOverhead is the header overhead for IPv6 based connections.
const IPv6HeaderOverhead = 40

//UDPv4MinHeaderOverhead is the minimum header overhead for UDP based connections over IPv4.
const UDPv4MinHeaderOverhead = 28

//UDPv4MaxHeaderOverhead is the maximum header overhead for UDP based connections over IPv4.
const UDPv4MaxHeaderOverhead = 68

//UDPv6HeaderOverhead is the header overhead for UDP based connections over Ipv6.
const UDPv6HeaderOverhead = 48
