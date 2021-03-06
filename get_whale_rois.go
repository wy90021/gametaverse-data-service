package main

import (
	"gametaverse-data-service/schema"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetWhaleRois(fromTimeObj time.Time, toTimeObj time.Time, sortType schema.WhalesSortType) []schema.UserRoiDetail {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	})

	svc := s3.New(sess)

	newUsers := getNewUsers(fromTimeObj, toTimeObj, *svc)

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
	perNewUserRoiDetail := map[string]*schema.UserRoiDetail{}

	priceHistory := getPriceHistory("sea", fromTimeObj, toTimeObj, *svc)
	priceHisoryMap := map[int64]float64{}
	layout := "2006-01-02"
	for _, price := range priceHistory.Prices {
		timeObj, _ := time.Parse(layout, price.Date)
		priceHisoryMap[timeObj.Unix()] = price.Price
	}
	for _, transfer := range totalTransfers {
		//log.Printf("user %s transfer %v", "0xfff5de86577b3f778ac6cc236384ed6db1825bff", transfer)
		if joinedTimestamp, ok := newUsers[transfer.FromAddress]; ok {
			dateTimestamp := (int64(transfer.Timestamp) / int64(schema.DayInSec)) * int64(schema.DayInSec)
			valueUsd := (transfer.Value / float64(schema.SeaTokenUnit)) * priceHisoryMap[dateTimestamp]
			valueToken := transfer.Value / float64(schema.SeaTokenUnit)
			if userRoiDetails, ok := perNewUserRoiDetail[transfer.FromAddress]; ok {
				userRoiDetails.TotalProfitUsd -= valueUsd
				userRoiDetails.TotalSpendingUsd += valueUsd
				userRoiDetails.TotalProfitToken -= valueToken
				userRoiDetails.TotalSpendingToken += valueToken
			} else {
				perNewUserRoiDetail[transfer.FromAddress] = &schema.UserRoiDetail{
					UserAddress:        transfer.FromAddress,
					JoinDateTimestamp:  joinedTimestamp,
					TotalSpendingUsd:   valueUsd,
					TotalProfitUsd:     -valueUsd,
					TotalSpendingToken: valueToken,
					TotalProfitToken:   -valueToken,
					TotalGainUsd:       0,
					TotalGainToken:     0,
				}
			}
		}
		if joinedTimestamp, ok := newUsers[transfer.ToAddress]; ok {
			dateTimestamp := (int64(transfer.Timestamp) / int64(schema.DayInSec)) * int64(schema.DayInSec)
			valueUsd := (transfer.Value / float64(schema.SeaTokenUnit)) * priceHisoryMap[dateTimestamp]
			valueToken := transfer.Value / float64(schema.SeaTokenUnit)
			if userRoiDetails, ok := perNewUserRoiDetail[transfer.ToAddress]; ok {
				userRoiDetails.TotalProfitUsd += valueUsd
				userRoiDetails.TotalGainUsd += valueUsd
				userRoiDetails.TotalGainToken += valueToken
				userRoiDetails.TotalProfitToken += valueToken
			} else {
				perNewUserRoiDetail[transfer.ToAddress] = &schema.UserRoiDetail{
					UserAddress:        transfer.ToAddress,
					JoinDateTimestamp:  joinedTimestamp,
					TotalSpendingUsd:   0,
					TotalSpendingToken: 0,
					TotalGainUsd:       valueUsd,
					TotalGainToken:     valueToken,
					TotalProfitUsd:     valueUsd,
					TotalProfitToken:   valueToken,
				}
			}
		}
	}

	userRoiDetails := make([]schema.UserRoiDetail, len(perNewUserRoiDetail))
	profitableUserCount := 0
	idx := 0
	for _, userRoiDetail := range perNewUserRoiDetail {
		userRoiDetails[idx] = *userRoiDetail
		idx += 1
		if userRoiDetail.TotalProfitUsd > 0 {
			profitableUserCount += 1
		}
	}

	if sortType == schema.SortByGain {
		sort.Slice(userRoiDetails, func(i, j int) bool {
			return userRoiDetails[i].TotalGainToken > userRoiDetails[j].TotalGainToken
		})
	} else if sortType == schema.SortByProfit {
		sort.Slice(userRoiDetails, func(i, j int) bool {
			return userRoiDetails[i].TotalProfitToken > userRoiDetails[j].TotalProfitToken
		})
	} else if sortType == schema.SortBySpending {
		sort.Slice(userRoiDetails, func(i, j int) bool {
			return userRoiDetails[i].TotalSpendingToken > userRoiDetails[j].TotalSpendingToken
		})
	}
	return userRoiDetails[0:10]
}
