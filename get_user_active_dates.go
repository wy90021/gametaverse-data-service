package main

import (
	"gametaverse-data-service/schema"
	"math"
	"sort"
	"time"
)

func GetUserActiveDates(fromTimeObj time.Time, toTimeObj time.Time, limit int64) []schema.UserActivity {

	totalTransfers := GetTransfers(fromTimeObj, toTimeObj)
	//for _, item := range resp.Contents {
	//	log.Printf("file name: %s\n", *item.Key)
	//	timestamp, _ := strconv.ParseInt(strings.Split(*item.Key, "-")[0], 10, 64)
	//	timeObj := time.Unix(timestamp, 0)
	//	if timeObj.Before(fromTimeObj) || timeObj.After(toTimeObj) {
	//		continue
	//	}

	//	requestInput :=
	//		&s3.GetObjectInput{
	//			Bucket: aws.String(dailyTransferBucketName),
	//			Key:    aws.String(*item.Key),
	//		}
	//	result, err := svc.GetObject(requestInput)
	//	if err != nil {
	//		exitErrorf("Unable to get object, %v", err)
	//	}
	//	body, err := ioutil.ReadAll(result.Body)
	//	if err != nil {
	//		exitErrorf("Unable to get body, %v", err)
	//	}
	//	bodyString := string(body)
	//	//transactions := converCsvStringToTransactionStructs(bodyString)
	//	transfers := ConvertCsvStringToTransferStructs(bodyString)
	//	log.Printf("transfer num: %d", len(transfers))
	//	//dateString := time.Unix(int64(dateTimestamp), 0).UTC().Format("2006-January-01")
	//	totalTransfers = append(totalTransfers, transfers...)
	//}
	perUserTransfers := map[string][]schema.Transfer{}
	for _, transfer := range totalTransfers {
		if _, ok := perUserTransfers[transfer.FromAddress]; ok {
			perUserTransfers[transfer.FromAddress] = append(perUserTransfers[transfer.FromAddress], transfer)
		} else {
			perUserTransfers[transfer.FromAddress] = make([]schema.Transfer, 0)
		}
		if _, ok := perUserTransfers[transfer.ToAddress]; ok {
			perUserTransfers[transfer.ToAddress] = append(perUserTransfers[transfer.ToAddress], transfer)
		} else {
			perUserTransfers[transfer.ToAddress] = make([]schema.Transfer, 0)
		}
	}
	perUserActivities := make([]schema.UserActivity, 0) //len(perUserTransfers))
	idx := 0
	for userAddress, transfers := range perUserTransfers {
		if len(transfers) == 0 {
			continue
		}
		sort.Slice(transfers, func(i, j int) bool {
			return transfers[i].Timestamp < transfers[j].Timestamp
		})
		totalDatesCount := transfers[len(transfers)-1].Timestamp/schema.DayInSec - transfers[0].Timestamp/schema.DayInSec + 1
		activeDatesCount := 1
		currentDate := transfers[0].Timestamp / schema.DayInSec
		for _, transfer := range transfers {
			if transfer.Timestamp/schema.DayInSec != currentDate {
				activeDatesCount += 1
				currentDate = transfer.Timestamp / schema.DayInSec
			}
		}

		//if userAddress == "0x27eafaf87860c290c185c1105cbedeb3b742c748" {
		//	log.Printf("for user %s, totalDatesCount %d, activeDatesCount %d", userAddress, totalDatesCount, activeDatesCount)
		//	for _, transfer := range transfers {
		//		log.Printf("transfer timestamp %d, date %d", transfer.Timestamp, transfer.Timestamp/dayInSec)
		//	}
		//	perUserActivities[idx] = UserActivity{UserAddress: userAddress, TotalDatesCount: int64(totalDatesCount), ActiveDatesCount: int64(activeDatesCount)}
		//}
		perUserActivities = append(perUserActivities, schema.UserActivity{UserAddress: userAddress, TotalDatesCount: int64(totalDatesCount), ActiveDatesCount: int64(activeDatesCount)})
		idx += 1
	}
	sort.Slice(perUserActivities, func(i, j int) bool {
		return perUserActivities[i].TotalDatesCount > perUserActivities[j].TotalDatesCount
	})
	return perUserActivities[0:int64(math.Min(float64(limit), float64(len(perUserActivities))))]
}
