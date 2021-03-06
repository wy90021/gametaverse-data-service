package main

import (
	"context"
	"encoding/json"
	daily "gametaverse-data-service/daily/functions"
	"gametaverse-data-service/grafana"
	"gametaverse-data-service/lib"
	"gametaverse-data-service/schema"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type handler struct {
	s3Client *s3.S3
	cache    *lib.Cache
}

func (h *handler) process(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	input := schema.Input{}
	json.Unmarshal([]byte(request.Body), &input)
	log.Printf("path: %s, body: %s, httpmethod: %s", request.Path, request.Body, request.HTTPMethod)
	log.Printf("request: %v", request)
	log.Printf("Input: %v", input)
	if request.Path == "/grafana/search" {
		response := grafana.Search()
		return GenerateResponse(response)
	} else if request.Path == "/grafana/query" {
		grafanaQueryRequest := schema.GrafanaQueryRequest{}
		json.Unmarshal([]byte(request.Body), &grafanaQueryRequest)
		log.Printf("grafana/query body: %s", request.Body)
		log.Printf("grafana/query request: %v", grafanaQueryRequest)
		layout := "2006-01-02T15:04:05.000Z"
		fromTimeObj, _ := time.Parse(layout, grafanaQueryRequest.Range.From)
		toTimeObj, _ := time.Parse(layout, grafanaQueryRequest.Range.To)
		fromTimeDateObj := time.Unix((fromTimeObj.Unix()/int64(schema.DayInSec))*int64(schema.DayInSec), 0)
		toTimeDateObj := time.Unix((toTimeObj.Unix()/int64(schema.DayInSec))*int64(schema.DayInSec), 0)
		// Warm up cache
		if strings.HasSuffix(grafanaQueryRequest.Targets[0].Target, "2") {
			go lib.GetUserActionsRangeAsync(h.s3Client, h.cache, fromTimeObj.Unix(), toTimeObj.Unix())
		}
		if grafanaQueryRequest.Targets[0].Target == "daus" {
			log.Printf("grafana/query request from %v, to %v", fromTimeDateObj, toTimeDateObj)
			daus := GetGameDaus(fromTimeDateObj, toTimeDateObj)
			response := grafana.GetDauMetrics(daus)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "daily_transaction_volume" {
			dailyTransactionVolumes := GetGameDailyTransactionVolumes(fromTimeDateObj, toTimeDateObj)
			response := grafana.GetDailyTransactionVolumeMetrics(dailyTransactionVolumes)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_profitable_rate" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, toTimeDateObj, false)
			response := grafana.GetNewUserProfitableRateMetrics(newUserProfitableRate.OverallProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "user_repurchase_rate" {
			userRepurchaseRate := GetUserRepurchaseRate(fromTimeDateObj, toTimeDateObj)
			response := grafana.GetUserRepurchaseRateMetrics(userRepurchaseRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "user_actual_active_dates_distribution" {
			userActiveDates := GetUserActiveDates(fromTimeDateObj, toTimeDateObj, 10000000)
			response := grafana.GetUserActualActiveDatesDistributionMetrics(userActiveDates)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "user_total_active_dates_distribution" {
			userActiveDates := GetUserActiveDates(fromTimeDateObj, toTimeDateObj, 10000000)
			response := grafana.GetUserTotalActiveDatesDistributionMetrics(userActiveDates)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_spending_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewUserSpendingUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_spending_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewRenteeSpendingUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_spending_token_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewRenteeSpendingTokenDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_spending_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewPurchaserSpendingUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_spending_token_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewPurchaserSpendingTokenDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_spending_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewHybriderSpendingUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_spending_token_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewHybriderSpendingTokenDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_profit_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewUserProfitUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_profit_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewRenteeProfitUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_profit_token_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewRenteeProfitTokenDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_profit_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewPurchaserProfitUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_profit_token_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewPurchaserProfitTokenDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_profit_usd_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewHybriderProfitUsdDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_profit_token_distribution" {
			newUserProfitableRate := GetNewUserProfitableRate(fromTimeDateObj, time.Now(), true)
			response := grafana.GetNewHybriderProfitTokenDistributionMetrics(newUserProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_type" {
			newUserTypes := GetUserType(fromTimeDateObj, time.Now())
			response := grafana.GetNewUserTypeMetrics(newUserTypes)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_profitable_days" {
			userRois := GetUserRoi(fromTimeDateObj, time.Now())
			response := grafana.GetNewUserProfitableDaysDistributionMetrics(userRois)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_profitable_days" {
			userRois := GetUserRoi(fromTimeDateObj, time.Now())
			response := grafana.GetNewRenteeProfitableDaysDistributionMetrics(userRois)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_profitable_days" {
			userRois := GetUserRoi(fromTimeDateObj, time.Now())
			response := grafana.GetNewPurchaserProfitableDaysDistributionMetrics(userRois)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_profitable_days" {
			userRois := GetUserRoi(fromTimeDateObj, time.Now())
			response := grafana.GetNewHybriderProfitableDaysDistributionMetrics(userRois)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "whale_sort_by_gain" {
			whaleRois := GetWhaleRois(schema.StarSharksStartingDate, time.Now(), schema.SortByGain)
			response := grafana.GetWhaleRoisMetrics(whaleRois, schema.SortByGain)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "whale_sort_by_profit" {
			whaleRois := GetWhaleRois(schema.StarSharksStartingDate, time.Now(), schema.SortByProfit)
			response := grafana.GetWhaleRoisMetrics(whaleRois, schema.SortByProfit)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "whale_sort_by_spending" {
			whaleRois := GetWhaleRois(schema.StarSharksStartingDate, time.Now(), schema.SortBySpending)
			response := grafana.GetWhaleRoisMetrics(whaleRois, schema.SortBySpending)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "daus2" {
			daus := daily.GetDaus(h.s3Client, h.cache, fromTimeDateObj, toTimeDateObj)
			response := grafana.GetDauMetrics(daus)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "daily_transaction_volume2" {
			dailyTransactionVolumes := daily.GetTransactionVolumes(h.s3Client, h.cache, fromTimeObj, toTimeObj)
			response := grafana.GetDailyTransactionVolumeMetrics(dailyTransactionVolumes)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "user_repurchase_rate2" {
			userRepurchaseRate := daily.GetUserRepurchaseRate(h.s3Client, h.cache, fromTimeObj.Unix(), toTimeObj.Unix())
			response := grafana.GetUserRepurchaseRateMetrics(userRepurchaseRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "user_actual_active_dates_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), toTimeObj.Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "user_actual_active_dates_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			userActiveDates := daily.GetUserActiveDays(h.s3Client, h.cache, fromTimeObj.Unix(), toTimeObj.Unix(), 1000000)
			response := grafana.GetUserActualActiveDatesDistributionMetrics(userActiveDates)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "user_actual_active_dates_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "user_total_active_dates_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), toTimeObj.Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "user_total_active_dates_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			userActiveDates := daily.GetUserActiveDays(h.s3Client, h.cache, fromTimeObj.Unix(), toTimeObj.Unix(), 1000000)
			response := grafana.GetUserTotalActiveDatesDistributionMetrics(userActiveDates)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "user_total_active_dates_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_type2" {
			newUserTypes := daily.GetUserType(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix())
			response := grafana.GetNewUserTypeMetrics(newUserTypes)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_profitable_rate2" {
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), toTimeObj.Unix(), false, fromTimeObj, toTimeObj)
			response := grafana.GetNewUserProfitableRateMetrics(newUserProfitableRate.OverallProfitableRate)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_spending_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_user_spending_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewUserSpendingUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_user_spending_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_spending_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_rentee_spending_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewRenteeSpendingUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_rentee_spending_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_spending_token_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_rentee_spending_token_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewRenteeSpendingTokenDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_rentee_spending_token_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_spending_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_purchaser_spending_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewPurchaserSpendingUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_purchaser_spending_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_spending_token_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_purchaser_spending_token_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewPurchaserSpendingTokenDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_purchaser_spending_token_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_spending_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_hybrider_spending_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewHybriderSpendingUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_hybrider_spending_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_spending_token_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_hybrider_spending_token_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewHybriderSpendingTokenDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_hybrider_spending_token_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_profit_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_user_profit_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewUserProfitUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_user_profit_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_profit_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_rentee_profit_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewRenteeProfitUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_rentee_profit_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_profit_token_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_rentee_profit_token_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewRenteeProfitTokenDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_rentee_profit_token_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_profit_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_purchaser_profit_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewPurchaserProfitUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_purchaser_profit_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_profit_token_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_purchaser_profit_token_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewPurchaserProfitTokenDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_purchaser_profit_token_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_profit_usd_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_hybrider_profit_usd_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewHybriderProfitUsdDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_hybrider_profit_usd_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_profit_token_distribution2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_hybrider_profit_token_distribution2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			newUserProfitableRate := daily.GetNewUserProfitableRate(h.s3Client, h.cache, fromTimeObj.Unix(), time.Now().Unix(), true, fromTimeObj, toTimeObj)
			response := grafana.GetNewHybriderProfitTokenDistributionMetrics(newUserProfitableRate)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_hybrider_profit_token_distribution2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_user_profitable_days2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_user_profitable_days2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			userRois := daily.GetNewUserRoi(h.s3Client, h.cache, fromTimeObj, time.Now())
			response := grafana.GetNewUserProfitableDaysDistributionMetrics(userRois)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_user_profitable_days2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_rentee_profitable_days2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_rentee_profitable_days2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			userRois := daily.GetNewUserRoi(h.s3Client, h.cache, fromTimeObj, time.Now())
			response := grafana.GetNewRenteeProfitableDaysDistributionMetrics(userRois)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_rentee_profitable_days2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_purchaser_profitable_days2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_purchaser_profitable_days2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			userRois := daily.GetNewUserRoi(h.s3Client, h.cache, fromTimeObj, time.Now())
			response := grafana.GetNewPurchaserProfitableDaysDistributionMetrics(userRois)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_purchaser_profitable_days2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "new_hybrider_profitable_days2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "new_hybrider_profitable_days2"); exist {
				json.Unmarshal(body, &resp)
				return GenerateResponse(resp)
			}
			userRois := daily.GetNewUserRoi(h.s3Client, h.cache, fromTimeObj, time.Now())
			response := grafana.GetNewHybriderProfitableDaysDistributionMetrics(userRois)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "new_hybrider_profitable_days2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "whale_sort_by_gain2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "whale_sort_by_gain2"); exist {
				if err := json.Unmarshal(body, &resp); err != nil {
					return GenerateResponse(resp)
				}
			}
			whaleRois := daily.GetWhaleRois(h.s3Client, h.cache, schema.StarSharksStartingDate.Unix(), time.Now().Unix(), schema.SortByGain)
			response := grafana.GetWhaleRoisMetrics(whaleRois, schema.SortByGain)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "whale_sort_by_gain2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "whale_sort_by_profit2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "whale_sort_by_profit2"); exist {
				if err := json.Unmarshal(body, &resp); err != nil {
					return GenerateResponse(resp)
				}
			}
			whaleRois := daily.GetWhaleRois(h.s3Client, h.cache, schema.StarSharksStartingDate.Unix(), time.Now().Unix(), schema.SortByProfit)
			response := grafana.GetWhaleRoisMetrics(whaleRois, schema.SortByProfit)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "whale_sort_by_profit2", resByte)
			return GenerateResponse(response)
		} else if grafanaQueryRequest.Targets[0].Target == "whale_sort_by_spending2" {
			var resp grafana.QueryResponse
			key := lib.GetDateRange(fromTimeObj.Unix(), time.Now().Unix())
			if body, exist := lib.GetRangeCacheFromS3(h.s3Client, key, "whale_sort_by_spending2"); exist {
				if err := json.Unmarshal(body, &resp); err != nil {
					return GenerateResponse(resp)
				}
			}
			whaleRois := daily.GetWhaleRois(h.s3Client, h.cache, schema.StarSharksStartingDate.Unix(), time.Now().Unix(), schema.SortBySpending)
			response := grafana.GetWhaleRoisMetrics(whaleRois, schema.SortBySpending)
			resByte, _ := json.Marshal(response)
			go lib.SetRangeCacheFromS3(h.s3Client, key, "whale_sort_by_spending2", resByte)
			return GenerateResponse(response)
		}
		return GenerateResponse("")
	} else if input.Method == "getDaus" {
		response := GetGameDaus(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0))
		return GenerateResponse(response)
	} else if input.Method == "getDailyTransactionVolumes" {
		response := GetGameDailyTransactionVolumes(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0))
		return GenerateResponse(response)
		//} else if input.Method == "getUserData" {
		//	response := getUserData(input.Params[0].Address)
		//	return GenerateResponse(response)
	} else if input.Method == "getUserRetentionRate" {
		response := GetUserRetentionRate(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0))
		return GenerateResponse(response)
	} else if input.Method == "getUserRepurchaseRate" {
		response := GetUserRepurchaseRate(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0))
		return GenerateResponse(response)
	} else if input.Method == "getUserSpendingDistribution" {
		response := GetUserSpendingDistribution(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0))
		return GenerateResponse(response)
	} else if input.Method == "getUserProfitDistribution" {
		userAddressMap := map[string]bool{}
		for _, param := range input.Params {
			userAddressMap[param.Address] = true
		}
		response := GetUserProfitDistribution(userAddressMap)
		return GenerateResponse(response)
		//return generateJsonResponse(response)
	} else if input.Method == "getUserRoi" {
		response := GetUserRoi(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0))
		return GenerateResponse(response)
		//return generateJsonResponse(response)
	} else if input.Method == "getUserActiveDates" {
		response := GetUserActiveDates(schema.StarSharksStartingDate, time.Now(), input.Params[0].Limit)
		return GenerateResponse(response)
		//return generateJsonResponse(response)
	} else if input.Method == "getNewUserProfitableRate" {
		response := GetNewUserProfitableRate(time.Unix(input.Params[0].FromTimestamp, 0), time.Now(), false)
		return GenerateResponse(response)
	} else if input.Method == "getNewUserProfitableRateDebug" {
		response := GetNewUserProfitableRate(time.Unix(input.Params[0].FromTimestamp, 0), time.Unix(input.Params[0].ToTimestamp, 0), true)
		return GenerateResponse(response)
	} else if input.Method == "getUserType" {
		fromTimeObj := schema.StarSharksStartingDate
		toTimeObj := time.Now()
		response := GetUserType(fromTimeObj, toTimeObj)
		return GenerateResponse(response)
	} else if input.Method == "getWhaleRois" {
		fromTimeObj := schema.StarSharksStartingDate
		toTimeObj := time.Now()
		response := GetWhaleRois(fromTimeObj, toTimeObj, schema.SortByGain)
		return GenerateResponse(response)
	} else if input.Method == "getUserActiveDays" {
		response := daily.GetUserActiveDays(h.s3Client, h.cache, input.Params[0].FromTimestamp, input.Params[0].ToTimestamp, 1000000)
		return GenerateResponse(response)
	}
	return GenerateResponse("")
}

func main() {
	sess, _ := session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				Region: aws.String("us-west-1"),
			},
		},
	)
	s3client := s3.New(sess)
	h := handler{
		s3Client: s3client,
		cache:    lib.NewCache(),
	}
	lambda.Start(h.process)
}
