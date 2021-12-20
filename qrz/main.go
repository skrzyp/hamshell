package qrz

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pterm/pterm"
)

type Callsign struct {
	XMLName   xml.Name `xml:"QRZDatabase"`
	Callsign  string   `xml:"Callsign>call"`
	FirstName string   `xml:"Callsign>fname"`
	LastName  string   `xml:"Callsign>name"`
	City      string   `xml:"Callsign>addr2"`
	Country   string   `xml:"Callsign>country"`
}

func (c *Callsign) Print() {
	pterm.DefaultTable.WithData(pterm.TableData{
		{"callsign", c.Callsign},
		{"first name", c.FirstName},
		{"last name", c.LastName},
		{"city", c.City},
		{"country", c.Country},
	}).WithLeftAlignment().Render()
}

type Session struct {
	loginQRZ    string
	passwordQRZ string

	tokenQRZ       string
	tokenQRZexpiry time.Time

	client *http.Client
}

type tokendataQRZ struct {
	XMLName      xml.Name `xml:"QRZDatabase"`
	Key          string   `xml:"Session>Key"`
	Subscription string   `xml:"Session>SubExp"`
}

func (s *Session) getToken() error {
	resp, err := s.client.PostForm("https://xmldata.qrz.com/xml/current/",
		url.Values{
			"username": {s.loginQRZ},
			"password": {s.passwordQRZ},
			"agent":    {"hamshell"},
		})
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	tokendata := tokendataQRZ{}
	err = xml.Unmarshal(body, &tokendata)
	if err != nil {
		return err
	}
	if tokendata.Key == "" {
		return fmt.Errorf("error retrieving qrz.com key from xml session data")
	}

	s.tokenQRZ = tokendata.Key
	s.tokenQRZexpiry = time.Now().Add(time.Hour * 1)

	return nil
}

func (s *Session) GetCall(callstr string) (call Callsign, err error) {
	if time.Now().After(s.tokenQRZexpiry) {
		err := s.getToken()
		if err != nil {
			return call, err
		}
	}
	resp, err := s.client.PostForm("https://xmldata.qrz.com/xml/current/",
		url.Values{
			"s":        {s.tokenQRZ},
			"callsign": {callstr},
			"agent":    {"hamshell"},
		})
	defer resp.Body.Close()
	if err != nil {
		return call, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return call, err
	}
	err = xml.Unmarshal(body, &call)
	if err != nil {
		return call, err
	}
	return call, nil
}

func New(loginQRZ string, passwordQRZ string) (s *Session, err error) {
	s = &Session{
		loginQRZ:    loginQRZ,
		passwordQRZ: passwordQRZ,
		client:      &http.Client{},
	}
	err = s.getToken()
	if err != nil {
		return s, fmt.Errorf("error getting token: %w", err)
	}
	return s, nil
}
