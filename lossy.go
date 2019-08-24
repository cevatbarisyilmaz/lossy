/*

Package lossy simulates bandwidth, latency and packet loss for net.PacketConn and net.Conn interfaces.

*/
package lossy

//Minimum Header Overhead for IPv4 based connections
const IPv4MinHeaderOverhead = 20

//Maximum Header Overhead for IPv4 based connections
const IPv4MaxHeaderOverhead = 60

//Header Overhead for IPv6 based connections
const IPv6HeaderOverhead = 40

//Minimum Header Overhead for UDP based connections over IPv4
const UDPv4MinHeaderOverhead = 28

//Maximum Header Overhead for UDP based connections over IPv4
const UDPv4MaxHeaderOverhead = 68

//Header Overhead for UDP based connections over Ipv6
const UDPv6HeaderOverhead = 48
