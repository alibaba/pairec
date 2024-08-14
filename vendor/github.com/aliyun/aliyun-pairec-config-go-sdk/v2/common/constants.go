package common

import "errors"

const (
	// Environment value list
	Environment_Daily   = 1
	Environment_Prepub  = 2
	Environment_Product = 4

	Environment_Daily_Desc   = "daily"
	Environment_Prepub_Desc  = "prepub"
	Environment_Product_Desc = "product"

	Bucket_Type_UID      = uint32(1)
	Bucket_Type_UID_HASH = uint32(2)
	Bucket_Type_Custom   = uint32(3)
	Bucket_Type_Filter   = uint32(4)

	ExpRoom_Status_Offline = uint32(1)
	ExpRoom_Status_Online  = uint32(2)

	ExpRoom_Type_Base   = uint32(1)
	ExpRoom_Type_Normal = uint32(2)

	Experiment_Status_Offline = uint32(1)
	Experiment_Status_Online  = uint32(2)

	Experiment_Type_Base    = uint32(1)
	Experiment_Type_Test    = uint32(2)
	Experiment_Type_Default = uint32(3)

	ExpGroup_Status_Offline                 = uint32(1)
	ExpGroup_Status_Online                  = uint32(2)
	ExpGroup_Status_Push_ALL                = uint32(4)
	ExpGroup_Distribution_Type_User         = 1
	ExpGroup_Distribution_Type_TimeDuration = 2

	Need_Feature_Reply_No  = 1
	Need_Feature_Reply_Yes = 2

	Feature_Consistency_Job_State_NO_RUN  = 1
	Feature_Consistency_Job_State_RUNNING = 2

	CODE_OK                            = "OK"
	Feature_Consistency_Job_Param_Name = "_feature_consistency_job_"
)

var (
	Environment2string = map[int]string{
		Environment_Daily:   "daily",
		Environment_Prepub:  "prepub",
		Environment_Product: "product",
	}
	EnvironmentDesc2OpenApiString = map[string]string{
		Environment_Daily_Desc:   "Daily",
		Environment_Prepub_Desc:  "Pre",
		Environment_Product_Desc: "Prod",
	}
	OpenapiEnvironment2Environment = map[string]int{
		"Daily": Environment_Daily,
		"Pre":   Environment_Prepub,
		"Prod":  Environment_Product,
	}
)

func CheckEnvironmentValue(env string) error {
	found := false
	for _, str := range Environment2string {
		if str == env {
			found = true
			break
		}
	}

	if !found {
		return errors.New("invalid environment value:" + env)
	}
	return nil
}
