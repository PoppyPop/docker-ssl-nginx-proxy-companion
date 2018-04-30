package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"bufio"
	. "github.com/PoppyPop/cfssl-go-client"
	"time"
)

func HandleDomain(domain []string) (missing []string, renew []*JsonInfoResponse) {
	var renewCheck []string

	for _, element := range domain {
		if _, err := os.Stat(*sslDir + element + ".crt"); os.IsNotExist(err) {
			missing = append(missing, element)
		} else {
			renewCheck = append(renewCheck, element)
		}
	}

	// check server for certificate renew
	s := NewServer(*serverUrl)
	for _, element := range renewCheck {
		buf, _ := ioutil.ReadFile(*sslDir + element + ".crt")
		crtString := string(buf)

		var rq JsonInfoRequest
		rq.Certificate = crtString
		res, err := s.CertInfo(rq)

		if res != nil && err == nil {
			// 720 hours = 1 month
			if res.NotAfter.Sub(time.Now()).Hours() < 720 {
				renew = append(renew, res)
			}
		}
	}

	return missing, renew
}

func CreateCert(domains []string) {
	if len(domains) > 0 {
		fmt.Printf("[%s]Generating cert\n", domains[0])

		s := NewServer(*serverUrl)

		var rq JsonNewCertRequest

		rq.Request.CN = domains[0]
		rq.Request.Hosts = domains
		rq.Profile = Profile

		newcert, err := s.NewCert(rq)

		if newcert != nil && err == nil {
			for _, element := range domains {
				err := ioutil.WriteFile(*sslDir+element+".crt", []byte(newcert.Certificate), 0644)
				if err != nil {
					fmt.Println(err)
				}

				err = ioutil.WriteFile(*sslDir+element+".csr", []byte(newcert.CertificateRequest), 0644)
				if err != nil {
					fmt.Println(err)
				}

				err = ioutil.WriteFile(*sslDir+element+".key", []byte(newcert.PrivateKey), 0644)
				if err != nil {
					fmt.Println(err)
				}
			}

			fmt.Printf("[%s]Success\n", domains[0])
		} else {
			fmt.Printf("[%s]Error generating cert: %s\n", domains[0], err)
		}
	}
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func RenewCert(certs []*JsonInfoResponse) {
	//s := NewServer(*serverUrl)

	var todo map[string][]string
	todo = make(map[string][]string)

	for _, cert := range certs {
		todo[cert.Subject.CommonName] = unique(append(todo[cert.Subject.CommonName], cert.SubjectAlternativeNames...))
		todo[cert.Subject.CommonName] = deleteSlice(todo[cert.Subject.CommonName], cert.Subject.CommonName)
	}

	if len(todo) > 0 {
		s := NewServer(*serverUrl)

		for key, value := range todo {

			fmt.Printf("[%s]Renewing cert\n", key)
			buf, _ := ioutil.ReadFile(*sslDir + key + ".csr")
			csrString := string(buf)

			var rq JsonSignRequest
			rq.Profile = Profile
			rq.Request = csrString

			crt, err := s.Sign(rq)

			if crt != nil || err == nil {
				ReplaceCert(key, crt)

				for _, san := range value {
					ReplaceCert(san, crt)
				}
				fmt.Printf("[%s]Success\n", key)
			} else {
				fmt.Printf("[%s]Error renewing cert: %s\n", key, err.Message)
			}
		}
	}
}

func ReplaceCert(name string, crt []byte) {
	os.Remove(*sslDir + name + ".old.crt")
	os.Rename(*sslDir+name+".crt", *sslDir+name+".old.crt")

	err := ioutil.WriteFile(*sslDir+name+".crt", crt, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func pos(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func deleteSlice(slice []string, value string) []string {
	i := pos(slice, value)
	if i > -1 {
		slice[i] = slice[len(slice)-1] // Replace it with the last one.
		slice = slice[:len(slice)-1]   // Chop off the last one.
	}
	return slice
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

var serverUrl = flag.String("server", "http://yugo.moot.fr:8888", "server url")
var sslDir = flag.String("certs", "/certs/", "certs directory")
var hostFile = flag.String("file", "/cfssl_service_data", "hosts file")

const Profile = "server"

func main() {
	flag.Parse()
	log.SetFlags(0)

	if _, err := os.Stat(*hostFile); !os.IsNotExist(err) {
		hosts, err := readLines(*hostFile)

		if err == nil {
			missing, renew := HandleDomain(hosts)

			CreateCert(missing)

			RenewCert(renew)
		} else {
			fmt.Printf("Unable to read hosts file: %s\n", *hostFile)
		}
	} else {
		fmt.Printf("Missing hosts file: %s\n", *hostFile)
	}
}
