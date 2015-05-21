package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apcera/termtables"
	"github.com/ttacon/chalk"
	"github.com/ziutek/mymysql/autorc"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

func console(db *autorc.Conn) {
	fmt.Printf("mysql [%s] > ", chalk.Bold.TextStyle(currentDB(db)))
	scanner := bufio.NewScanner(os.Stdin)
	lines := make([]string, 128)
	for c := 0; scanner.Scan() && c < len(lines); c++ {
		lines[c] = scanner.Text()
		if strings.Contains(lines[c], ";") {
			break
		}
		fmt.Printf(" > ")
	}
	sqlquery := strings.Join(lines, " ")

	if !db.Raw.IsConnected() {
		fmt.Println(chalk.Yellow, "Reconnecting to DB...", chalk.Reset)
	}

	start := time.Now()
	rows, res, err := db.Query(sqlquery)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Println(chalk.Red.Color(err.Error()))
		return
	}

	if len(rows) < 1 || res.StatusOnly() {
		return
	}

	displayResults(rows, res, elapsed)
}

func displayResults(rows []mysql.Row, res mysql.Result, elapsed time.Duration) {
	termtables.EnableUTF8PerLocale()
	table := termtables.CreateTable()

	fields := make([]interface{}, len(res.Fields()))
	for c, field := range res.Fields() {
		fields[c] = fmt.Sprintf("%s", field.Name)
	}
	table.AddHeaders(fields...)

	for _, row := range rows {
		columns := make([]interface{}, len(row))
		for c, col := range row {
			if col == nil {
				col = "NULL"
			}
			columns[c] = fmt.Sprintf("%s", col)
		}
		table.AddRow(columns...)
	}
	fmt.Print(table.Render())

	rowstr := "rows"
	if len(rows) == 1 {
		rowstr = "row"
	}
	fmt.Printf("%d %s in set (%s)\n", len(rows), rowstr, elapsed)
	if res.WarnCount() != 0 {
		fmt.Printf("%s%s warning(s).%s\n", chalk.Yellow, res.WarnCount(), chalk.Reset)
	}
	fmt.Println()
}

func main() {
	user, password, host, dbname := parseCmdLine()
	db := autorc.New("tcp", "", host+":3306", user, password, dbname)
	err := db.Raw.Connect()

	if err != nil {
		fmt.Println(chalk.Red.Color(err.Error()))
		flag.Usage()
		os.Exit(1)
	}

	defer db.Raw.Close()
	for {
		console(db)
	}
}

func currentDB(db *autorc.Conn) string {
	sql := "select database()"
	rows, _, err := db.Query(sql)
	if err != nil {
		fmt.Println(chalk.Red.Color(err.Error()))
		os.Exit(1)
	}

	if rows[0][0] == nil {
		return "none"
	}

	return fmt.Sprintf("%s", rows[0][0])
}

func parseCmdLine() (user, pass, host, dbname string) {
	userPtr := flag.String("user", "root", "username")
	passPtr := flag.String("password", "", "password")
	hostPtr := flag.String("host", "localhost", "hostname")
	dbPtr := flag.String("db", "", "database name")
	flag.Parse()

	user = string(*userPtr)
	pass = string(*passPtr)
	host = string(*hostPtr)
	dbname = string(*dbPtr)
	return
}
