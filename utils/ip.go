package utils

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"net/netip"
	"strings"
)

func GetClientIp(ctx *fiber.Ctx, all ...bool) string {
	var ips []string
	ips = append(ips, ctx.Get("X-Real-IP"))
	ips = append(ips, ctx.IPs()...)
	ips = append(ips, ctx.IP())

	ips = UniqueList(ips)

	log.Println("[request]", string(ctx.Request().RequestURI()), strings.Join(ips, ", "))

	if len(all) > 0 && all[0] {
		return strings.Join(ips, ", ")
	}

	for _, ip := range ips {
		addr, err := netip.ParseAddr(ip)
		if err == nil && !addr.IsPrivate() {
			return ip
		}
	}
	if len(ips) > 0 {
		return strings.Join(ips, ", ")
	}

	return ""
}
