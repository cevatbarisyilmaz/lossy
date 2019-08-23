/*

Package lossy simulates bandwidth, latency and packet loss for net.PacketConn and net.Conn interfaces.

It can be used to test robustness of applications and network protocols runs over unreliable protocols such as UDP or IP.

*/
package lossy

const plainUDPHeaderOverHead = 8

const IPv4MinHeaderOverhead = 20
const IPv4MaxHeaderOverhead = 60
const IPv6HeaderOverhead = 40
const UDPv4MinHeaderOverhead = IPv4MinHeaderOverhead + plainUDPHeaderOverHead
const UDPv4MaxHeaderOverhead = IPv4MaxHeaderOverhead + plainUDPHeaderOverHead
const UDPv6HeaderOverhead = IPv6HeaderOverhead + plainUDPHeaderOverHead
