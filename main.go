// PSNbot project main.go
package main

import (
	"PSNapi/handlers"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var usrnm = flag.String("username", "jiegokoji", "")
var fast = flag.String("fast", "yes", "")

type games struct {
	TitleId           string `json:"titleId"`
	NpCommunicationId string `json:"NpCommunicationId"`
}

func main() {
	os.Setenv("HTTP_PROXY", "")
	flag.Parse()
	var err error
	db, err = sql.Open("mysql", "root:_pass_@/psn")
	if err != nil {
		fmt.Println(err.Error())
	}
	//defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}
	if *fast == "no" {
		FillUser("f00c5319-2325-4eb7-b0a5-fe15a09fd44d", *usrnm)
		FillMessages("f00c5319-2325-4eb7-b0a5-fe15a09fd44d", *usrnm)
		FillGames("f00c5319-2325-4eb7-b0a5-fe15a09fd44d", *usrnm)

	} else {
		if *fast == "yes" {
			FillUser("f00c5319-2325-4eb7-b0a5-fe15a09fd44d", *usrnm)
			FillMessages("f00c5319-2325-4eb7-b0a5-fe15a09fd44d", *usrnm)
		}
		if *fast == "friend" {
			addfriend("f00c5319-2325-4eb7-b0a5-fe15a09fd44d", *usrnm)
		}

	}
	//fmt.Println("Hello World!")
}
func FillMessages(refreshtoken string, username string) {
	oauth, err := handlers.Login(refreshtoken)
	if err != nil {
		fmt.Println(err)
	}
	threads, _ := handlers.MessageThreads(oauth, username)
	for _, title := range threads.ThreadIds {
		threadInfo, _ := handlers.MessageThreadInfo(oauth, title.ThreadId)
		for _, events := range threadInfo.ThreadEvents {
			fmt.Println(events.MessageEventDetail.MessageDetail.Body)
			fmt.Println(events.MessageEventDetail.PostDate)
			if len(events.MessageEventDetail.AttachedMediaPath) > 1 {
				now := time.Now() // current local time
				uni := now.Unix()
				rows, _ := db.Query("SELECT mpath FROM `user_msg` WHERE mpath = '" + events.MessageEventDetail.AttachedMediaPath + "';")
				//fmt.Println(err)
				var mpath string
				for rows.Next() {
					err = rows.Scan(&mpath)
				}
				if len(mpath) > 1 {

				} else {
					handlers.MessageAttachment(oauth, events.MessageEventDetail.AttachedMediaPath, "/home/www/images/"+username+"_"+strconv.FormatInt(uni, 10))
					db.Exec("INSERT IGNORE INTO `user_msg` (`onlineId`, `threaId`, `message`, `attach`, `postdate`, `send`, `mpath`) VALUES ('" + username + "', '" + title.ThreadId + "', '" + events.MessageEventDetail.MessageDetail.Body + "', 'images/" + username + "_" + strconv.FormatInt(uni, 10) + ".jpg" + "', '" + events.MessageEventDetail.PostDate + "', '0','" + events.MessageEventDetail.AttachedMediaPath + "')")
				}

			}

		}
	}
}
func FillUser(refreshtoken string, username string) {
	oauth, err := handlers.Login(refreshtoken)
	if err != nil {
		fmt.Println(err, oauth)
	}
	profile, err := handlers.UserInfo(oauth, username)
	if err != nil {
		fmt.Println(err)
	}
	games, err := handlers.UserGames(oauth, username, "1", "0")
	if err != nil {
		fmt.Println(err)
	}
	db.Exec("REPLACE INTO `users` (`onlineId`, `lastOnline`, `plus`, `lastPlayed`, `platinum`, `gold`, `silver`, `bronze`, `totalgames`) VALUES ('" + username + "', '" + games.TrophyTitles[0].ComparedUser.LastUpdateDate + "', '" + strconv.Itoa(profile.Profile.Plus) + "', '" + games.TrophyTitles[0].NpCommunicationId + "', '" + strconv.Itoa(profile.Profile.TrophySummary.EarnedTrophies.Platinum) + "', '" + strconv.Itoa(profile.Profile.TrophySummary.EarnedTrophies.Gold) + "', '" + strconv.Itoa(profile.Profile.TrophySummary.EarnedTrophies.Silver) + "', '" + strconv.Itoa(profile.Profile.TrophySummary.EarnedTrophies.Bronze) + "','" + strconv.Itoa(games.TotalResults) + "')")
	var t2 time.Time
	timeNow := time.Now().UTC().Add(time.Minute * -time.Duration(15)).Format(time.RFC3339)
	if *fast == "no" {
		timeNow = time.Now().UTC().Add(time.Minute * -time.Duration(440)).Format(time.RFC3339)
	}
	t1, err := time.Parse(time.RFC3339, timeNow)
	if err != nil {
		fmt.Println(err)
		return
	}

	alltrophies, err := handlers.GetGameTrophies(oauth, games.TrophyTitles[0].NpCommunicationId, username)
	alltrophies_rar, err := handlers.GetGameTrophieData(oauth, games.TrophyTitles[0].NpCommunicationId, "default") // handlers.GetGameTrophies(oauth, games.TrophyTitles[0].NpCommunicationId, "")
	if err != nil {
		fmt.Println(err)
	}

	for _, trophie := range alltrophies.Trophies {
		if len(trophie.ComparedUser.EarnedDate) > 2 {
			t2, err = time.Parse(time.RFC3339, trophie.ComparedUser.EarnedDate)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if trophie.ComparedUser.OnlineId != "" && trophie.ComparedUser.Earned == true && inTimeSpan(t1, t2) {
			fmt.Println(trophie.TrophyName)
			fmt.Println(trophie.ComparedUser.EarnedDate)
			fmt.Println(trophie.TrophyType)
			fmt.Println(trophie.TrophyDetail)
			fmt.Println(trophie.TrophyIconUrl)
			fmt.Println(trophie.TrophyId)
			for _, trophie_rar := range alltrophies_rar.Trophies {
				if trophie_rar.TrophyId == trophie.TrophyId {
					fmt.Println("earn rate ", trophie.TrophyName)
					fmt.Println("earn rate ", trophie_rar.TrophyEarnedRate)
					//res, err := db.Exec("INSERT IGNORE INTO `trophyqueue` (`gameId`, `trophyName`, `trophyType`, `trophyDetail`, `trophyIcon`, `onlineId`, `earnedDate`, `send`, `earnrate`) VALUES ('" + games.TrophyTitles[0].NpCommunicationId + "', '" + MysqlRealEscapeString(trophie.TrophyName) + "', '" + trophie.TrophyType + "', '" + MysqlRealEscapeString(trophie.TrophyDetail) + "', '" + trophie.TrophyIconUrl + "', '" + username + "', '" + trophie.ComparedUser.EarnedDate + "', '0', '" + trophie_rar.TrophyEarnedRate + "')")
					res, err := db.Exec("INSERT IGNORE INTO `trophyqueue` (`gameId`, `trophyName`, `trophyType`, `trophyDetail`, `trophyIcon`, `onlineId`, `earnedDate`, `send`, `earnrate`) VALUES (?,?,?,?,?,?,?,?,?)", games.TrophyTitles[0].NpCommunicationId, trophie.TrophyName, trophie.TrophyType, MysqlRealEscapeString(trophie.TrophyDetail), trophie.TrophyIconUrl, username, trophie.ComparedUser.EarnedDate, "0", trophie_rar.TrophyEarnedRate)
					fmt.Println(res, err)
				}
			}

		}
	}
	fmt.Println(profile.Profile.Plus)
}
func FillGames(refreshtoken string, username string) {
	oauth, err := handlers.Login(refreshtoken)
	if err != nil {
		fmt.Println(err)
	}
	games, _ := handlers.UserGames(oauth, username, "1", "0")
	total := games.TotalResults
	offset := 0
	fmt.Println(total)
	for total > 100 {
		games, err = handlers.UserGames(oauth, username, "100", strconv.Itoa(offset))
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, title := range games.TrophyTitles {
			//fmt.Println(title.NpCommunicationId)
			//fmt.Println(title.TrophyTitleName)
			//fmt.Println(title.TrophyTitleIconUrl)
			//fmt.Println(title.TrophyTitlePlatfrom)
			//res, err := db.Exec("INSERT IGNORE INTO `games` (`name`, `Image`, `NpCommunicationId`, `platform`) VALUES ('" + MysqlRealEscapeString(title.TrophyTitleName) + "', '" + title.TrophyTitleIconUrl + "', '" + title.NpCommunicationId + "', '" + title.TrophyTitlePlatfrom + "')")
			res, err := db.Exec("INSERT IGNORE INTO `games` (`name`, `Image`, `NpCommunicationId`, `platform`) VALUES (?,?,?,?)", title.TrophyTitleName, title.TrophyTitleIconUrl, title.NpCommunicationId, title.TrophyTitlePlatfrom)
			fmt.Println(res, " NpCommunicationId ", err)
			db.Exec("REPLACE INTO `user_earnings` (`onlieId`, `gameId`, `progress`,`Platinum`,`lastUpdate`) VALUES ('" + title.ComparedUser.OnlineId + "', '" + title.NpCommunicationId + "', '" + strconv.Itoa(title.ComparedUser.Progress) + "', '" + strconv.Itoa(title.ComparedUser.EarnedTrophies.Platinum) + "', '" + title.ComparedUser.LastUpdateDate + "')")
		}
		total = total - 100
		offset = offset + 100
		fmt.Println("total left:", total, " offset: ", offset)
	}
	if total != 0 {
		games, err = handlers.UserGames(oauth, username, "100", strconv.Itoa(offset))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	for _, title := range games.TrophyTitles {
		//fmt.Println(title.NpCommunicationId)
		//fmt.Println(title.TrophyTitleName)
		//fmt.Println(title.TrophyTitleIconUrl)
		//fmt.Println(title.TrophyTitlePlatfrom)
		db.Exec("INSERT IGNORE INTO `games` (`name`, `Image`, `NpCommunicationId`, `platform`) VALUES ('" + MysqlRealEscapeString(title.TrophyTitleName) + "', '" + title.TrophyTitleIconUrl + "', '" + title.NpCommunicationId + "', '" + title.TrophyTitlePlatfrom + "')")
		//fmt.Println(res, " NpCommunicationId ", err)
		db.Exec("REPLACE INTO `user_earnings` (`onlieId`, `gameId`, `progress`,`Platinum`,`lastUpdate`) VALUES ('" + title.ComparedUser.OnlineId + "', '" + title.NpCommunicationId + "', '" + strconv.Itoa(title.ComparedUser.Progress) + "', '" + strconv.Itoa(title.ComparedUser.EarnedTrophies.Platinum) + "', '" + title.ComparedUser.LastUpdateDate + "')")
	}

}
func addfriend(refreshtoken string, psnid string) {
	oauth, err := handlers.Login(refreshtoken)
	if err != nil {
		fmt.Println(err)
	}
	handlers.UserAddFriend(oauth, "ledokol322", psnid, "test msg")
	res, err := db.Exec("INSERT IGNORE INTO `friends`  (`psn`) VALUES (?)", psnid)
	fmt.Println(res, err)
}
func MysqlRealEscapeString(value string) string {
	replace := map[string]string{"\\": "\\\\", "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z", "®": "", "'": `\'`}

	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}

	return value
}
func inTimeSpan(start, check time.Time) bool {
	//fmt.Println(check.After(start), "  ", check.Before(end))
	return check.After(start)
}
