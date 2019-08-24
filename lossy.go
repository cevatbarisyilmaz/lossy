/*

Package lossy simulates bandwidth, latency and packet loss for net.PacketConn and net.Conn interfaces.

*/
package lossy

const IPv4MinHeaderOverhead = 20
const IPv4MaxHeaderOverhead = 60
const IPv6HeaderOverhead = 40
const UDPv4MinHeaderOverhead = 28
const UDPv4MaxHeaderOverhead = 68
const UDPv6HeaderOverhead = 48
