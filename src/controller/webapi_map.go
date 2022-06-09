package controller

import (
	"fmt"
	"sync"
	"time"

	"mocaApi/src/controller/cls"
)

type AgentInfo struct {
	uuid     string
	devId    int
	siteId   int
	regionId int
	manager  int

	pubIP   string
	sysInfo string
	//	ispInfo  string //이전버전
	//	sysState string //이전버전
	danger bool

	nicInfo string
	plgInfo string

	command string

	fireVer int
	defVer  int

	floodTime   int
	floodCnt    int
	fBlockTime  int
	readTime    int
	readMinSize int
	rBlockTime  int

	setTime uint

	idleDomain []string
	delBlack   []string
	danCount   int  //추가사항
	danTime    uint //추가사항
}

type PlgInfo struct {
	version int
	kind    int
}

type SiteInfo struct {
	inTime     int
	plgString  string
	certString string
	plgList    map[int]PlgInfo
	certList   map[int]int
}

type RegionInfo struct {
	sphere  string
	station string
	db      string
}

var agentMap = struct {
	sync.RWMutex
	data map[int]AgentInfo
}{data: make(map[int]AgentInfo)}

var siteMap = struct {
	sync.RWMutex
	data map[int]SiteInfo
}{data: make(map[int]SiteInfo)}

var regionMap = struct {
	sync.RWMutex
	data map[int]RegionInfo
}{data: make(map[int]RegionInfo)}

func SetSiteMap(siteId int, siteInfo SiteInfo) {
	siteMap.Lock()
	siteMap.data[siteId] = siteInfo
	siteMap.Unlock()
}

func GetSiteMap(siteId int) (SiteInfo, bool) {
	siteMap.RLock()
	sInfo, ok := siteMap.data[siteId]
	siteMap.RUnlock()

	return sInfo, ok
}

func GetSiteMapTTL(siteId, ttl int) (SiteInfo, bool) {
	siteMap.RLock()
	sInfo, ok := siteMap.data[siteId]
	siteMap.RUnlock()

	nowTime := int(time.Now().Unix())
	if ok && sInfo.inTime+ttl < nowTime {
		sInfo.inTime = nowTime - ttl + 2
		SetSiteMap(siteId, sInfo)
		return sInfo, false
	}

	return sInfo, ok
}

func GetAgentMap(devid int) (AgentInfo, bool) {
	agentMap.RLock()
	agentInfo, ok := agentMap.data[devid]
	agentMap.RUnlock()

	return agentInfo, ok
}

func SetAgentMap(devid int, agentInfo AgentInfo) {
	agentMap.Lock()
	agentMap.data[devid] = agentInfo
	agentMap.Unlock()
}

func SetRegionMap(rid int, rInfo RegionInfo) {
	regionMap.Lock()
	regionMap.data[rid] = rInfo
	regionMap.Unlock()
}

func GetRegionMap(rid int) (RegionInfo, bool) {
	regionMap.RLock()
	rInfo, ok := regionMap.data[rid]
	regionMap.RUnlock()

	return rInfo, ok
}

func LoadAgent() {
	var devid int
	var ip string

	query := fmt.Sprintf("SELECT DEVID, IP FROM Agent_TAB")
	lprintf(4, "[IFNO] select query : %s\n", query)

	rows, err := cls.SqliteDB.Query(query)
	if err != nil {
		lprintf(1, "[ERROR] select error : %s\n", err.Error())
		return
	}
	defer rows.Close()

	//nowTime := int(time.Now().Unix())
	for rows.Next() { // whole row
		if err = rows.Scan(&devid, &ip); err != nil {
			lprintf(1, "[ERROR] sql scan error(%s)\n", err.Error())
		}

		var aInfo AgentInfo
		aInfo.pubIP = ip
		aInfo.devId = devid
		lprintf(4, "[INFO] insert agent from sql result (%d:%s)\n", devid, ip)
		SetAgentMap(devid, aInfo)
	}

	go AliveAgentCheck()
}

func AliveAgentCheck() {
	count := 0
	time.Sleep(60 * time.Second)
	for {
		time.Sleep(10 * time.Second)
		var delList []int
		nTime := uint(time.Now().Unix())
		agentMap.RLock()
		for _, aInfo := range agentMap.data {
			if aInfo.setTime+60 > nTime {
				continue
			}

			// 1분간 보고 없는 녀석
			lprintf(4, "[INFO] this node(%d):(%s) did not report for 60 second \n", aInfo.devId, aInfo.pubIP)
			//			if aInfo.devId != 0 && sendNotiAgent(aInfo, 3, "") != "200" {
			// fail (it might be down or chaged other device - new device id)
			//			if setAgentState(aInfo.devId, true) == false {
			//			continue
			//	}
			// }
			delList = append(delList, aInfo.devId)
		}
		agentMap.RUnlock()

		// remove
		for i := 0; i < len(delList); i++ {
			deleteAgentLite(delList[i])
			agentMap.Lock()
			delete(agentMap.data, delList[i])
			agentMap.Unlock()
		}

		// status
		if count == 3 {
			count = 0
			var agentList string
			agentMap.RLock()
			for _, aInfo := range agentMap.data {
				agentList += fmt.Sprintf("%d(%s)-", aInfo.devId, aInfo.pubIP)
			}
			agentCnt := len(agentMap.data)
			agentMap.RUnlock()

			statData := fmt.Sprintf("name:smartsphere,regionid:%d,agentCnt:%d,agentList:%s", 1 /*MyRegionID*/, agentCnt, agentList)

			/*
				url := "http://127.0.0.1:" + AgentNotifyPort + "/smartagent/stat"
				req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte(statData)))
				req.Header.Set("Connection", "close")
				client := &http.Client{}
				resp, err := client.Do(req)
			*/
			resp, err := cls.HttpRequestDetail("HTTP", "GET", "127.0.0.1", "1111" /*AgentNotifyPort*/, "smartagent/stat", []byte(statData), nil, "", true)
			if err != nil {
				lprintf(1, "[FAIL] noti request error (%s)", err)
				continue
			}
			resp.Body.Close()
		}
		count++
	}
}

func deleteAgentLite(devid int) bool {

	// use
	query := fmt.Sprintf("DELETE FROM AGENT_TAB WHERE DEVID = ?")
	lprintf(4, "[IFNO] update query : %s -> %d\n", query, devid)

	statement, err := cls.SqliteDB.Prepare(query)
	if err != nil {
		lprintf(1, "[ERROR] sql prepare error : %s\n", err.Error())
		return false
	}
	defer statement.Close()

	_, err = statement.Exec(devid)
	if err != nil {
		lprintf(1, "[ERROR] sql insert exec error : %s\n", err.Error())
		return false
	}
	return true
}

func updateAgentLite(devid int, ip string) bool {

	// use
	nowTime := int(time.Now().Unix())

	query := fmt.Sprintf("INSERT OR REPLACE INTO AGENT_TAB(DEVID, IP, INTIME) VALUES (?, ?, ?)")
	lprintf(4, "[IFNO] update query : %s -> %d, %s\n", query, devid, ip)

	statement, err := cls.SqliteDB.Prepare(query)
	if err != nil {
		lprintf(1, "[ERROR] sql prepare error : %s\n", err.Error())
		return false
	}
	defer statement.Close()

	_, err = statement.Exec(devid, ip, nowTime)
	if err != nil {
		lprintf(1, "[ERROR] sql insert exec error : %s\n", err.Error())
		return false
	}
	return true
}
