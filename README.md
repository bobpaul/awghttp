# awghttp

Turn WireGuard to HTTP & SOCKS5 proxies.

Supports [AmneziaWG protocol extensions](https://docs.amnezia.org/documentation/amnezia-wg/) to hide from Deep Packet Inspection firewalls.
A more complete description of AmneziaWG's obfuscation strategy can be found here](https://mk16.de/blog/amneziawg-en/)

Based on the work of [zhsj's wghttp](https://github.com/zhsj/wghttp)

The HTTP & SOCKS5 proxies are served on the same port. It runs in userspace,
without requiring a WireGuard kernel module or TUN device.

In remote exit mode, the proxy is served on local network, and the traffic
from proxy server goes to WireGuard network.

In local exit mode, the proxy is served on WireGuard network, and the traffic
from WireGuard goes to local network.

For detailed usage, see <https://github.com/bobpaul/awghttp/tree/master/docs>.
