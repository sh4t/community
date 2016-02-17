package host

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)
// Definte what a host is:
type Sensor struct {
	Name			string	`json:"name"`
	Ports			[]int	`json:"ports"`
}
type Host struct {
	Id       		bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt		time.Time	`json:"created"`
	ModifiedAt		time.Time	`json:"modified"`
	Hostname     	string	`json:"hostname"`
	HostType		string	`json:"type"`
	HostOs			string	`json:"os"`
	HostArch		string	`json:"architecture"`

	Specs			struct	{
		CpuCount	string	`json:"cpu_count"`
		CpuFreq		string	`json:"cpu_freq"`
		Memory		string	`json:"memory"`
		Storage		string	`json:"storage"`
		DiskType	string	`json:"disk_type"`
		Hypervisor	string	`json:"hypervisor"`
	} `json:"resources"`
	
	Ips 			struct	{
		Ipv4		string	`json:"primary_ipv4"`
		Ipv6		string	`json:"primary_ipv6,omitempty"`
		AddIpv4		[]string	`json:"ipv4"`
		AddIpv6		[]string	`json:"ipv6"`
	} `json:"ip_addresses"`

	Provider			struct	{
		Name 			string	`json:"name"`
		Website			string	`json:"website,omitempty"`
	} `json:"provider"`

	Sensors			[]Sensor `json:"sensors"`
}

type HostsCollection struct {
	Data []Host `json:"data"`
}

type HostResource struct {
	Data Host `json:"data"`
}