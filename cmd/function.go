package main

func getIP(ipport string, xrealip string, xffr string) string {
	if xrealip != "" {
		return xrealip
	} else if xffr != "" {
		return xffr
	} else {
		host, _, _ := net.SplitHostPort(ipport)
		return host
	}
}
