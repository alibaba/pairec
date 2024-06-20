package web

import (
	"encoding/json"
	"reflect"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

type param struct {
	ComplexTypeFeatures ComplexTypeFeatures `json:"complex_type_features"`
	Uid                 string              `json:"uid"`
}

func TestComplexTypeFeatures(t *testing.T) {

	test_cases := []Feature{
		{
			Name:   "test_list_int",
			Values: []int{1, 2, 3},
			Type:   "list<int>",
		},
		{
			Name:   "test_int64",
			Values: int64(456),
			Type:   "int64",
		},
		{
			Name:   "test_int",
			Values: int(34567),
			Type:   "int",
		},
		{
			Name:   "test_float",
			Values: float32(0.45),
			Type:   "float",
		},
		{
			Name:   "test_double",
			Values: float64(0.45),
			Type:   "double",
		},
		{
			Name:   "test_list_float",
			Values: []float32{1, 2, 3},
			Type:   "list<float>",
		},
		{
			Name:   "test_list_double",
			Values: []float64{0.1, 2, 0.3},
			Type:   "list<double>",
		},
		{
			Name:   "test_list_string",
			Values: []string{"a", "b", "c"},
			Type:   "list<string>",
		},
		{
			Name:   "test_map_string_int",
			Values: map[string]int{"a": 1, "b": 2, "c": 3},
			Type:   "map<string,int>",
		},
		{
			Name:   "test_map_string_double",
			Values: map[string]float64{"a": 1, "b": 2, "c": 3},
			Type:   "map<string,double>",
		},
		{
			Name:   "test_map_int_int",
			Values: map[int]int{11: 1, 22: 2, 33: 3},
			Type:   "map<int,int>",
		},
		{
			Name:   "test_map_int_int64",
			Values: map[int]int64{11: 1, 22: 2, 33: 3},
			Type:   "map<int,int64>",
		},
		{
			Name:   "test_map_int64_float",
			Values: map[int64]float32{11: 0.1, 22: 0.2, 33: 0.3},
			Type:   "map<int64,float>",
		},
		{
			Name:   "test_map_int64_float",
			Values: map[int64]float32{11: 0.1, 22: 0.2, 33: 0.3},
			Type:   "map<int64,float>",
		},
		{
			Name:   "test_map_int64_string",
			Values: map[int64]string{11: "0.1", 22: "0.2", 33: "0.3"},
			Type:   "map<int64,string>",
		},
	}
	m := make(map[string]any)
	m["uid"] = "123"
	m["complex_type_features"] = test_cases

	data, _ := json.Marshal(m)
	t.Log(string(data))
	//var data = `{"uid":"123", "complex_type_features":[{"name":"test_list_int", "values":[1,2,3], "type":"list<int>"}, {"name":"test_int64", "values":123, "type":"int64"}]}`

	param := param{}
	err := json.Unmarshal([]byte(data), &param)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range param.ComplexTypeFeatures.FeaturesMap {
		for _, c := range test_cases {
			if c.Name == k {
				if !reflect.DeepEqual(c.Values, v) {
					t.Fatalf("value not equal, expect:%v, actual:%v, expect type:%T, actual type:%T\n", c.Values, v, c.Values, v)
				}
				break
			}
		}
	}
	t.Log(param)
}

var data = `
	{"features":{"accountcreatetime":1677133264000,"accountid":"222539680","biztimestamp":1718263898768,"companytype":"02","is_distance":"1","lastnapplyjob":"","long_apply_context_workfunc_jobworkfunc_map":"0703:45928\u001d2302:7682\u001d2304:6097\u001d0711:4821\u001d2303:3819\u001d0602:1995\u001d0601:904\u001d3012:883\u001d0702:747\u001d0603:690\u001d3105:610\u001d4121:554\u001d2301:527\u001d2310:450\u001d0705:434\u001d0604:406\u001d1810:392\u001d3103:385\u001d0202:346\u001dA0K9:334\u001d4301:317\u001d0201:278\u001d0611:274\u001d0302:270\u001d3009:255\u001d4702:219\u001d4704:218\u001d2305:213\u001d3904:201\u001d3106:198\u001d0301:195\u001d3001:186\u001d4303:179\u001d0304:177\u001d0701:175\u001d8050:169\u001d8049:166\u001d2106:165\u001d3101:162\u001d2606:158\u001d8302:149\u001d0710:143\u001dA0L6:141\u001d3102:141\u001d0232:138\u001d3508:134\u001d3111:133\u001d2306:131\u001d4002:125\u001d3505:125\u001d0203:123\u001d0235:122\u001d3501:109\u001d3109:99\u001d0305:93\u001d1701:91\u001d0226:91\u001d2608:90\u001dA0LE:85\u001d4302:85\u001d0303:80\u001d4006:78\u001d3506:78\u001d4207:75\u001d3004:73\u001d0306:73\u001d0605:73\u001d0709:72\u001d1401:70\u001d0402:68\u001d3902:68\u001d4604:67\u001d8305:66\u001d0626:66\u001d4001:65\u001d0606:63\u001d2001:62\u001d3514:59\u001d4601:59\u001d0629:56\u001d0324:52\u001d3903:52\u001d4313:52\u001d8047:50\u001d3110:50\u001d3015:45\u001d0405:44\u001d1411:44\u001d5101:43\u001d1002:42\u001d3905:42\u001d3909:42\u001d3507:42\u001d8048:40\u001d4007:35\u001d0208:35\u001d0814:34\u001d0584:31\u001d0829:30\u001d0627:30","long_apply_user_job_companysize_map":"1:2\u001d2:1\u001d4:1","long_apply_user_job_companytype_map":"06:4","long_apply_user_job_degreefrom_map":"5:3\u001d6:1","long_apply_user_job_industrytype1_map":"03:1\u001d51:1\u001d04:1\u001d05:1","long_apply_user_job_jobsalarydown_map":"9:2\u001d7:1\u001d16:1","long_apply_user_job_jobsalaryup_map":"11:1\u001d13:1\u001d12:1\u001d20:1","long_apply_user_job_workfunc_map":"0703:1\u001d0711:1\u001d2302:1\u001d0629:1","long_apply_user_job_workyear_map":"5:2\u001d4:1\u001d6:1","long_apply_user_same_language_job_total_map":"0:4","long_apply_user_total_map":"4","long_chat_context_workfunc_jobworkfunc_map":"0703:11494\u001d2302:1681\u001d2304:1481\u001d0711:1195\u001d2303:791\u001d0602:376\u001d3105:202\u001d3012:197\u001d1810:170\u001d0601:153\u001d0702:139\u001d4121:133\u001d0603:105\u001d2301:96\u001d2310:87\u001d0705:87\u001d0604:83\u001d4301:82\u001d0202:81\u001d3103:79\u001dA0K9:73\u001d2305:69\u001d3009:63\u001d3001:55\u001d4303:54\u001d4704:50\u001d0302:48\u001d0304:41\u001d3106:38\u001d0701:37\u001d0305:35\u001d2606:33\u001d1701:31\u001d4302:30\u001d3101:29\u001d0203:27\u001d3904:27\u001d0235:27\u001d8049:25\u001dA0KF:25\u001d3111:25\u001d4002:24\u001d0710:24\u001d4702:24\u001d3508:24\u001d2306:23\u001d8302:23\u001d0605:23\u001d0611:23\u001d3004:22\u001d8047:22\u001d3902:22\u001d4604:21\u001d0301:21\u001d8050:21\u001d3102:21\u001d0402:20\u001dA0LE:20\u001d0405:19\u001d0201:19\u001d4006:19\u001d0226:19\u001d4207:18\u001d0303:18\u001d0232:18\u001d4901:17\u001d1401:17\u001d4601:16\u001d2608:16\u001d2106:16\u001d0626:16\u001d3110:16\u001d8305:15\u001d0306:15\u001d2610:15\u001d0324:15\u001dA0L6:15\u001d3109:15\u001d3505:15\u001d3501:15\u001d1411:14\u001d2001:14\u001dA0KE:14\u001d0629:14\u001d3905:14\u001d3506:14\u001d0606:13\u001d4007:13\u001d0404:12\u001d3015:12\u001d4001:12\u001d3507:12\u001d0403:11\u001d8301:11\u001d4112:11\u001d8203:11\u001d0584:11\u001d3903:11\u001d4306:10\u001d4703:10","long_chat_user_job_companysize_map":"1:2\u001d3:2\u001d2:2\u001d5:3\u001d4:2\u001d7:2","long_chat_user_job_companytype_map":"13:1\u001d06:13","long_chat_user_job_degreefrom_map":"5:7\u001d6:7","long_chat_user_job_industrytype1_map":"26:1\u001d01:1\u001d06:2\u001d58:2\u001d05:3\u001d46:1\u001d32:1\u001d23:1\u001d51:1\u001d09:1","long_chat_user_job_jobsalarydown_map":"11:1\u001d10:1\u001d13:1\u001d16:2\u001d7:3\u001d9:5\u001d8:1","long_chat_user_job_jobsalaryup_map":"11:4\u001d13:2\u001d12:2\u001d20:1\u001d17:1\u001d19:2\u001d18:1\u001d9:1","long_chat_user_job_workfunc_map":"0703:7\u001d0711:2\u001d2302:2\u001d2304:2\u001d0629:1","long_chat_user_job_workyear_map":"10:1\u001d3:3\u001d5:5\u001d4:2\u001d6:3","long_chat_user_same_language_job_total_map":"0:14","long_chat_user_total_map":"14","long_exposure_context_workfunc_jobworkfunc_map":"0703:4596695\u001d2302:574662\u001d2304:456354\u001d0711:388953\u001d2303:235630\u001d0602:152335\u001d3012:79375\u001d0702:58964\u001d2207:56964\u001d3105:55141\u001d0601:53078\u001d0603:50746\u001d4121:43195\u001d0202:39609\u001d3001:38384\u001d3103:36732\u001d2301:36396\u001d0705:36301\u001d1810:33074\u001d0604:31450\u001d4301:28667\u001d0302:27413\u001d3009:26474\u001dA0K9:21747\u001d2310:21596\u001d4704:21153\u001d0611:19435\u001d4303:19296\u001d0201:19009\u001d3106:17698\u001d4002:17337\u001d0304:17247\u001d3508:16562\u001d0701:16474\u001d4001:15706\u001d0226:15357\u001d4702:14887\u001d0232:13327\u001d0301:13142\u001d8049:13004\u001d2606:12930\u001d3101:12711\u001d0203:12367\u001d2305:12222\u001d8050:11369\u001d3904:11222\u001d3111:10979\u001d2001:10944\u001d8302:10856\u001d1701:10691\u001dA0L6:9869\u001d0235:9842\u001d3004:9692\u001d4302:9570\u001d3102:9399\u001d0303:9324\u001d3501:9210\u001d2306:9109\u001d0305:8832\u001d4006:8107\u001d3507:7909\u001d0709:7850\u001d0710:7556\u001d3505:7538\u001d3506:7094\u001d0626:7088\u001d3015:6188\u001d0584:6162\u001d2106:6087\u001d2608:6024\u001d0606:5904\u001d0402:5738\u001d0605:5670\u001d3109:5604\u001d1401:5404\u001d5101:5400\u001d3110:5261\u001d0629:5124\u001d0405:4970\u001d8047:4791\u001d4207:4640\u001d3514:4324\u001d8305:4120\u001d0306:4083\u001d3902:4077\u001d4112:4003\u001d3903:3960\u001d2610:3939\u001dA0KF:3933\u001d4313:3645\u001d0208:3470\u001dA0LE:3239\u001d0336:3165\u001d3601:3063\u001d8048:3044\u001d3014:2986\u001d4601:2822\u001d4604:2558\u001d5114:2507\u001d2002:2473","long_exposure_user_job_companysize_map":"1:241\u001d3:579\u001d2:595\u001d5:239\u001d4:258\u001d7:111\u001d6:19","long_exposure_user_job_companytype_map":"11:5\u001d02:83\u001d13:58\u001d01:14\u001d06:1840\u001d05:54\u001d03:59\u001d10:1","long_exposure_user_job_degreefrom_map":"1:4\u001d2:24\u001d5:1009\u001d4:40\u001d7:26\u001d6:991","long_exposure_user_job_industrytype1_map":"09:106\u001d37:1\u001d22:71\u001d36:10\u001d60:2\u001d61:39\u001d63:4\u001d65:9\u001d02:109\u001d03:41\u001d23:28\u001d01:85\u001d06:136\u001d07:26\u001d48:6\u001d05:471\u001d46:110\u001d47:27\u001d44:3\u001d45:2\u001d29:13\u001d40:1\u001d35:5\u001d14:63\u001d24:11\u001d56:25\u001d33:29\u001d43:7\u001d26:88\u001d20:16\u001d27:3\u001d59:1\u001d58:147\u001d11:16\u001d13:2\u001d12:11\u001d15:3\u001d04:82\u001d17:3\u001d38:6\u001d55:3\u001d54:7\u001d31:9\u001d49:37\u001d51:49\u001d50:27\u001d53:3\u001d19:23\u001d21:28\u001d32:64\u001d08:46","long_exposure_user_job_jobsalarydown_map":"11:333\u001d24:1\u001d13:104\u001d12:36\u001d15:4\u001d14:11\u001d22:3\u001d23:2\u001d19:61\u001d18:2\u001d16:160\u001d21:15\u001d10:62\u001d5:78\u001d25:4\u001d7:311\u001d6:129\u001d9:584\u001d8:184\u001d17:2\u001d4:28","long_exposure_user_job_jobsalaryup_map":"24:12\u001d25:12\u001d26:7\u001d20:32\u001d21:75\u001d22:1\u001d23:53\u001d5:12\u001d4:1\u001d7:49\u001d6:17\u001d9:251\u001d8:59\u001d11:437\u001d10:38\u001d13:263\u001d12:39\u001d15:25\u001d14:41\u001d17:50\u001d16:309\u001d19:240\u001d18:91","long_exposure_user_job_workfunc_map":"4303:12\u001d1305:1\u001d3717:1\u001d2302:152\u001d1301:2\u001d4302:1\u001d2305:8\u001d4309:1\u001d4301:3\u001dA0K1:1\u001d4411:1\u001d3326:2\u001d5506:4\u001d5507:1\u001d4124:1\u001d3726:1\u001d4306:1\u001d6007:1\u001d4704:10\u001dA0LE:6\u001d6009:2\u001d0232:5\u001d0306:1\u001d0302:11\u001d0301:1\u001d2910:1\u001d0236:3\u001d0840:1\u001d6010:2\u001d0929:1\u001d6601:1\u001d0539:1\u001d4135:3\u001d4402:2\u001d4403:1\u001d0602:6\u001d0603:10\u001d0607:1\u001d0604:9\u001d0337:2\u001d1322:2\u001d5013:2\u001d4207:3\u001d4204:1\u001d8703:2\u001d2301:4\u001d4809:2\u001d2303:22\u001d6513:6\u001d6514:2\u001d2304:194\u001d4125:2\u001d2306:4\u001d1335:2\u001d0202:15\u001d0203:37\u001d0201:2\u001d0446:1\u001d3108:2\u001d1407:2\u001d5003:1\u001d5001:1\u001d8044:1\u001d8048:1\u001d8049:1\u001d6104:4\u001d4118:1\u001d3508:1\u001d3509:3\u001d3012:13\u001d3501:1\u001d3505:1\u001d3014:3\u001d0511:1\u001d7421:1\u001dA0K3:1\u001d8050:2\u001d3614:1\u001d8312:1\u001d3807:1\u001dA0K9:2\u001d3004:3\u001d3604:8\u001d0611:3\u001d3001:29\u001d3002:4\u001d4108:2\u001d3009:8\u001d3008:1\u001dA0JQ:1\u001d5701:1\u001d5702:2\u001d7413:3\u001d0584:1\u001d2111:1\u001d7419:1\u001d7417:2\u001d3602:1\u001d3601:2\u001d2120:1\u001d0829:2\u001d5102:1\u001d4604:1\u001d5101:8\u001d5104:1\u001d3203:1\u001d0801:1\u001d3106:7\u001dA0KF:2\u001dA0KE:2\u001d2102:1\u001d2106:2\u001d3904:5\u001d3905:1\u001d5117:4\u001d3902:3\u001d3903:1\u001d8059:2\u001d3735:1\u001d3734:3\u001d0402:1\u001d0403:2\u001d0404:2\u001d0405:4\u001d2130:1\u001d3105:8\u001d4002:5\u001d0814:1\u001d6102:3\u001d3102:1\u001d3103:1\u001d1113:1\u001dA0LC:1\u001d0711:142\u001d0561:5\u001d0560:1\u001dA0L9:1\u001d4313:1\u001dA0L7:3\u001dA0L6:4\u001d0226:10\u001d5202:3\u001d4503:1\u001d1810:8\u001d6914:1\u001d0410:3\u001d4915:1\u001d0629:20\u001d0626:1\u001d0703:1121\u001d0702:3\u001d0628:1\u001d4121:1\u001d0705:3\u001d3507:3","long_exposure_user_job_workyear_map":"10:200\u001d1:12\u001d3:334\u001d5:740\u001d4:375\u001d7:49\u001d6:393\u001d8:11","long_exposure_user_same_language_job_total_map":"1:12\u001d0:2102","long_exposure_user_total_map":"2114","long_look_context_workfunc_jobworkfunc_map":"0703:399335\u001d2302:58567\u001d2304:49012\u001d0711:38525\u001d2303:29449\u001d0602:13818\u001d3012:8236\u001d3105:5902\u001d0702:5502\u001d0601:5266\u001d0603:4968\u001d4121:4134\u001d1810:4001\u001d2301:3564\u001d3103:3530\u001d0604:3454\u001d0705:3426\u001d0202:3111\u001dA0K9:2876\u001d2310:2768\u001d4301:2592\u001d0302:2430\u001d3001:2362\u001d3009:2345\u001d2305:1997\u001d0611:1876\u001d4303:1864\u001d4704:1827\u001d3106:1807\u001d0201:1726\u001d0304:1535\u001d3508:1371\u001d4702:1370\u001d3101:1353\u001d0701:1347\u001d3904:1317\u001d0301:1314\u001d4002:1288\u001d8049:1259\u001d2606:1219\u001d3111:1198\u001d2306:1170\u001d0232:1159\u001d8050:1148\u001d0235:1077\u001d1701:1058\u001d8302:1032\u001d3102:1031\u001d0203:978\u001dA0L6:954\u001d4001:948\u001d0226:907\u001d2001:899\u001d0303:888\u001d0710:887\u001d4006:845\u001d0305:835\u001d4302:785\u001d3501:782\u001d3505:757\u001d3109:711\u001d2608:700\u001d3506:692\u001d3004:686\u001d2106:683\u001d0605:651\u001d0626:615\u001d4207:566\u001d2207:556\u001d3110:554\u001d0709:552\u001d8047:542\u001d0402:532\u001d3902:524\u001d0629:515\u001d8305:475\u001d0606:475\u001d1401:468\u001d0584:463\u001d3015:460\u001d0306:423\u001dA0LE:423\u001d2610:418\u001d3507:401\u001d4604:364\u001d0405:348\u001d4601:345\u001d5101:341\u001dA0KF:340\u001d4007:335\u001d4313:323\u001d3903:322\u001d3514:315\u001d8048:309\u001d0609:303\u001d8203:289\u001d0208:274\u001d1002:269\u001d4714:263\u001d0336:251","long_look_user_job_companysize_map":"1:24\u001d3:67\u001d2:72\u001d5:36\u001d4:25\u001d7:17\u001d6:5","long_look_user_job_companytype_map":"02:11\u001d03:7\u001d13:8\u001d01:1\u001d06:224\u001d05:1","long_look_user_job_degreefrom_map":"5:133\u001d4:7\u001d6:112","long_look_user_job_industrytype1_map":"56:1\u001d61:6\u001d49:1\u001d02:20\u001d03:5\u001d26:11\u001d01:9\u001d06:24\u001d07:3\u001d04:7\u001d23:3\u001d46:20\u001d08:5\u001d09:13\u001d14:4\u001d38:1\u001d58:24\u001d11:1\u001d12:3\u001d22:6\u001d33:2\u001d32:7\u001d05:61\u001d51:9\u001d50:3\u001d19:1\u001d21:2","long_look_user_job_jobsalarydown_map":"11:28\u001d10:6\u001d13:13\u001d12:8\u001d14:1\u001d16:31\u001d19:4\u001d5:1\u001d7:30\u001d6:10\u001d9:93\u001d8:27","long_look_user_job_jobsalaryup_map":"11:73\u001d10:1\u001d13:30\u001d12:9\u001d20:9\u001d14:5\u001d17:6\u001d16:28\u001d19:33\u001d18:18\u001d23:1\u001d9:27\u001d21:12","long_look_user_job_workfunc_map":"3012:4\u001d0232:2\u001d2303:2\u001d2302:21\u001d2304:32\u001d0226:2\u001d3106:2\u001d3001:5\u001d0203:6\u001d0703:143\u001d0711:23\u001d0629:7\u001d3009:2\u001d0302:1","long_look_user_job_workyear_map":"10:22\u001d1:1\u001d3:39\u001d5:103\u001d4:39\u001d7:3\u001d6:45","long_look_user_same_language_job_total_map":"0:252","long_look_user_total_map":"252","modelcontextindustry1":"00","modelcontextindustry2":"","modelcontextindustry3":"","modelcontextindustrynew":"00","modelcontextsalarydown":10000,"modelcontextsalaryup":15000,"modelcontextworkfunc":"0703","modelcontextworkfuncnameembeddingweightedv2":"0.04881095,-0.22465408,0.28065285,-0.09645436,-0.16072841,-0.0976091,0.033957865,0.034560658,-0.17466058,-0.086896315,0.24596968,-0.08060291,-0.05147334,-0.004967162,-0.16295604,0.42543152,-0.13694607,0.022193702,-0.50012696,0.30615264,0.2121379,-0.0021834217,0.01316325,-0.107589394,-0.009218594,0.1389255,-0.014548734,-0.03471422,0.0650112,0.18625344,-0.11386113,0.11641062","modelcontextworkfuncnameembeddingweightedv3":"-0.14161423,-0.26610634,0.17278713,0.4172698,0.22936909,-0.002549798,0.012356481,0.10964216,0.044557344,-0.11272513,0.0592834,-0.08246623,0.1275374,-0.33027658,0.19400053,0.05315587,-0.37417114,0.12145555,-0.095130175,-0.05799091,0.1440832,0.05325402,0.2255048,-0.12093457,0.12630008,-0.28457832,-0.1440669,0.00673517,8.113532E-5,-0.0061744503,-0.13310598,0.24045962","modelskillwords__embeddingv3":"-0.23370577,-0.21057619,0.12451071,0.16582057,-0.002370126,-0.09596536,-0.16041936,0.25308028,-0.17585725,0.103265055,-0.055493575,-0.21964978,0.1559986,-0.15050639,-0.0943043,0.12380303,-0.13389096,0.45361444,0.0020103392,0.022773238,0.26655567,-0.04225363,0.07074557,-0.10667504,0.16567501,-0.35056335,-0.31242424,0.009356822,0.026288304,0.033166673,0.08614515,-0.15565656;-0.06625341,-0.1903784,0.34827524,0.22144628,0.07322963,-0.0613199,0.10704864,0.06937674,-0.1334492,0.10354047,0.13495314,-0.3631053,-0.016345236,-0.31001392,0.03367153,0.07454615,-0.3357825,0.21062726,-0.2581185,0.14340186,-0.0048145163,0.025513172,-0.13624738,-0.21853957,0.27512565,0.08577775,-0.089081034,-0.14840807,0.05672667,0.18988621,-0.09620934,0.095603995;-0.076834895,-0.19072533,0.31915742,0.35238898,0.033126336,-0.0475646,0.195133,0.15705104,-0.10432112,0.1631413,0.11946529,-0.23577797,-0.0025002444,-0.14728186,0.004829896,-0.12599468,-0.34158412,0.31101605,-0.25417754,0.065517165,-0.19224904,0.09368424,-0.09111782,0.08006564,0.2982004,0.18379788,-0.09611984,0.061922118,0.14886543,0.084822044,-0.012076988,0.112173505;-0.1771928,-0.102757804,0.13392104,0.22404103,-0.004729807,0.028427377,0.072547704,0.28448611,-0.19947952,0.15134163,0.35116607,-0.2610202,-0.1290845,-0.030631965,0.040630367,-0.105680056,-0.24851342,0.181281,-0.101325564,0.19770701,-0.26901218,-0.094539605,0.16979526,-0.21598922,0.053988203,0.16551016,-0.28918356,0.11402565,-0.14406285,-0.09733618,-0.15189168,0.19318901;0.14584424,-0.06763,0.12597641,0.18112396,-0.0845358,-0.09596047,-0.19735138,0.41247866,-0.25499964,-0.19537316,0.26262087,-0.23180416,0.09041728,-0.009968973,0.16333836,-0.08793712,0.03908835,0.1259671,-0.27951038,0.105949424,0.046534393,0.250716,0.37224427,-0.17928009,0.15723477,0.03274137,-0.11358394,-0.11813115,-0.14471799,0.03192509,0.10304357,0.06118653;0.07560589,0.17405505,0.059139904,0.18004441,0.11466546,-0.26259568,0.04566427,0.030477813,-0.34321448,0.026631542,0.27729824,-0.41400555,-0.058253147,0.035739,-0.25998837,-0.057210654,0.06613501,0.28869843,-0.10408497,0.2266587,0.17063518,0.15914415,0.16587062,-0.19396472,0.22367056,0.10905368,-0.18622498,0.05755288,0.14664291,-0.0071244785,0.079696745,0.0573706;-0.21178466,-0.21147369,0.26272798,0.019507904,0.24288212,0.062000148,-0.07505742,0.052995984,-0.061105143,-0.0072717653,0.15748717,-0.48027784,0.01926009,-0.18332069,-0.058801066,0.20673642,-0.20985046,0.33741358,-0.15380667,0.100523695,0.18128778,-0.06949827,0.05683448,-0.34092444,0.0857913,-0.09322267,-0.18885134,-0.03798782,0.13123016,0.069913246,0.0650881,-0.044780087;0.049354848,-0.18819398,0.41358864,0.37863788,-0.049179994,0.03822826,-0.0686711,0.13440634,-0.2289569,0.04599899,0.040575158,-0.20023397,-0.0657614,-0.2571594,0.07205339,0.017416034,-0.27969894,0.15944324,-0.23819076,0.14327334,-0.04430603,0.08469534,-0.1966555,-0.30087972,0.16844177,0.17521155,-0.12201813,-0.15790927,0.013667389,0.1378996,-0.105472125,0.018956726;-0.024341872,-0.07213797,0.27142966,0.2844336,-0.077604145,-0.32608125,-0.15395658,0.3910232,-0.028787984,-0.11374176,-3.4714874E-4,-0.21199755,0.07921513,0.063541844,0.053598177,0.27451938,0.056501403,0.064008385,-0.3403033,0.17493668,0.12805556,0.22710417,0.07793782,-0.09113301,0.17984925,0.038364127,-0.16859339,-0.21931134,-0.03762111,-0.20716333,0.06594869,0.091355994;-0.1013874,-0.07665859,0.22932011,0.46286604,-0.18324292,-0.27477714,-0.14203888,0.2816201,-0.18645534,-0.0017538554,0.08894575,-0.115938164,-0.080428265,-0.006606183,0.10903244,-0.07624249,0.011605482,0.12265866,-0.2259247,0.10627859,0.1528956,0.25596052,0.05757525,-0.27121997,0.2151306,0.16555376,-0.14169638,-0.21544561,-0.1476544,0.020117147,0.122425176,0.086300015;-0.12978193,-0.18119624,0.20720674,0.045743156,-0.17484626,-0.18648495,0.07012052,0.30590957,0.0053618206,-0.024702776,-0.0022996995,-0.06842905,-0.15081473,-0.13937734,-0.018562159,0.13983685,0.06090049,0.04662534,-0.3024204,0.17510109,-0.20737009,0.41294727,0.123454,-0.22361198,0.29703924,0.32990298,-0.08194873,-0.1653838,-0.06655136,0.12012722,0.1289268,-0.025798567;0.12529226,-0.30458766,0.20393993,0.212593,-0.043437257,-0.16348398,-0.1403084,0.40874946,-0.10152026,0.049592216,-0.21142685,-0.18417022,0.03454115,-0.020455914,-0.086454585,0.08013917,-0.15623063,0.28523916,-0.08971561,0.19849974,0.2197181,0.11583721,0.12393932,-0.3129188,-0.012099497,-0.27704608,-0.13622352,0.064861864,-0.18903928,0.1068159,0.068642564,0.052484322;0.08135484,0.07002285,0.04179911,0.11816103,0.16190726,-0.04128834,-0.1158284,0.37395242,-0.21521236,-0.03614341,0.066074714,-0.23186216,-0.14471617,0.10059753,-0.006699524,0.1466752,-0.0054448457,0.34802333,-0.21546887,0.106648594,0.14537437,0.25638354,0.34211972,-0.32435694,0.25595108,-0.09252805,0.038647838,0.07535627,0.16975006,-0.024655296,0.13898279,-0.12316637;0.05269707,0.056239393,0.0670763,0.16195004,0.100715294,-0.39790758,-0.011696203,0.037247557,-0.3736555,-0.008099078,0.126993,-0.2615842,0.0050691,-0.011321509,-0.17383464,-0.1387944,0.2764864,0.24548495,-0.18463086,0.22611049,0.20159018,0.16423321,0.11743385,-0.17186904,0.17166306,0.1097566,-0.31146026,-0.18906742,0.031593215,-0.086483024,0.00901683,-0.06140642;-0.067686915,-0.20454605,0.36182967,0.20526464,0.35795093,-0.051828805,-0.060268015,-0.0010334065,0.012113914,0.017843531,-0.008218231,-0.24925202,0.07538138,-0.17757301,-0.20691203,0.0305693,-0.4321621,0.34925327,-0.052928805,0.026498144,0.17239183,0.09519036,-0.010380058,-0.18238816,0.22187227,-0.056771215,-0.12975897,0.07047075,0.08462383,-0.014313172,-0.18861821,0.029331148;0.20187755,-0.1969945,0.035289027,0.08374135,-0.0914809,-0.3089357,-0.08084326,0.18507499,-0.13436012,0.105094075,0.19984269,-0.13809656,-0.12814133,0.10422075,-0.06294128,0.0971861,-0.06368708,-0.08973787,-0.05821598,0.27101493,0.24614127,0.35442597,0.29204217,0.23069577,-0.052739218,-0.102829196,-0.13481502,0.20367861,0.24312384,0.1875673,0.15429603,-0.18443248;-0.21178466,-0.21147369,0.26272798,0.019507904,0.24288212,0.062000148,-0.07505742,0.052995984,-0.061105143,-0.0072717653,0.15748717,-0.48027784,0.01926009,-0.18332069,-0.058801066,0.20673642,-0.20985046,0.33741358,-0.15380667,0.100523695,0.18128778,-0.06949827,0.05683448,-0.34092444,0.0857913,-0.09322267,-0.18885134,-0.03798782,0.13123016,0.069913246,0.0650881,-0.044780087;0.07560589,0.17405505,0.059139904,0.18004441,0.11466546,-0.26259568,0.04566427,0.030477813,-0.34321448,0.026631542,0.27729824,-0.41400555,-0.058253147,0.035739,-0.25998837,-0.057210654,0.06613501,0.28869843,-0.10408497,0.2266587,0.17063518,0.15914415,0.16587062,-0.19396472,0.22367056,0.10905368,-0.18622498,0.05755288,0.14664291,-0.0071244785,0.079696745,0.0573706;-0.031788662,-0.17421924,0.32862267,-0.13923417,0.3146882,-0.19463684,-0.14012031,0.24558048,0.081986055,0.060275204,-0.03926326,-0.3370886,0.08045553,-0.16474132,-0.10184247,0.2772752,-0.22822915,0.13342792,-0.2891982,0.13898788,-0.036058277,0.07841544,0.11579411,-0.16632119,0.23267235,-0.08883028,-0.19561064,-0.028386401,-0.07575962,-0.11586931,-0.093663566,0.16616705;0.07845252,0.13764882,-0.010674107,0.04537326,0.12024504,-0.22956488,-0.112588875,0.0944765,-0.31064335,-0.0725252,0.26597282,-0.25064155,-0.18148461,-0.010710089,0.0036074324,-0.018506942,-0.0037800802,0.34245068,-0.1852497,0.0640517,0.16405271,0.21527937,0.38540554,-0.27023676,0.2490121,-0.11566374,-0.06455767,0.20757139,0.17977999,-0.032476418,0.091171876,-0.01944998;-0.21178466,-0.21147369,0.26272798,0.019507904,0.24288212,0.062000148,-0.07505742,0.052995984,-0.061105143,-0.0072717653,0.15748717,-0.48027784,0.01926009,-0.18332069,-0.058801066,0.20673642,-0.20985046,0.33741358,-0.15380667,0.100523695,0.18128778,-0.06949827,0.05683448,-0.34092444,0.0857913,-0.09322267,-0.18885134,-0.03798782,0.13123016,0.069913246,0.0650881,-0.044780087;-0.12730452,-0.24384336,0.13353175,0.37747264,-0.09069588,-0.11745236,-0.06759442,0.48289105,-0.2379117,0.07054851,0.026657159,-0.18618657,0.007545594,0.04693312,0.046371616,0.053737026,-0.050247192,0.13843699,-0.23368345,0.13047351,0.16595939,0.08390968,0.26056436,-0.3099083,0.07250643,-0.23015921,-0.041091736,0.0029288887,-0.14125583,-0.1116963,0.06970704,0.104567;-0.12555973,-0.20942463,0.23887226,0.41301313,-0.119477734,-0.1799803,-0.076401316,0.47877803,-0.19462791,0.029042354,0.23887528,-0.0777147,0.07492229,-0.08666193,0.060334496,-0.094278574,0.028470756,0.013516197,-0.17469722,0.16673294,0.07632058,0.16818295,0.15342581,-0.33265254,0.10973633,-0.015378452,-0.08699772,-0.10610991,-0.1578869,-0.07578329,0.08651484,0.0815083;-0.15929244,0.10572003,-0.026213542,0.15052032,0.0022940326,-0.21448773,-0.06300417,0.39127472,-0.16952762,-0.07521634,0.09743006,-0.18120763,0.02061051,-0.024873616,-0.21872397,0.06384242,-0.10663106,0.40759972,-0.27880973,0.051102128,0.044751063,0.19305904,0.20134836,-0.210358,0.3478995,0.09817657,-0.12535866,0.1364721,0.17310105,0.06263011,0.078297526,0.14281218;0.043818083,0.08378806,0.2407548,-0.22100683,0.123821825,-0.26085117,-0.18233098,0.035554677,-0.17970791,-0.100569226,0.2404474,-0.26094705,-0.04062212,-0.106088966,-0.12377278,-0.06011476,0.06810586,0.06144369,-0.272036,0.23100908,0.03263779,0.21242264,0.13833047,-0.055457644,0.41442773,0.044346776,-0.08256207,0.25997287,-0.24246137,0.123531245,-0.0344091,0.19667232;-0.099462,-0.2036977,-0.018912146,0.24798195,0.04971796,0.16315024,0.0146689685,-0.27867514,-0.059756875,0.029377637,0.45125452,-0.009679792,0.06464264,-0.16948994,-0.31521502,0.10194819,-0.16290098,0.11672169,0.2555127,0.12532791,0.07488415,0.47027966,-0.02004597,0.06611242,0.058949165,0.093890704,-0.07318392,-0.042984568,0.15619767,-0.0014920462,-0.0042296713,0.17991546;0.029819459,-0.22039115,0.042005327,0.35698253,0.11276461,-0.15175772,-0.10805612,-0.16562286,-0.14954986,-0.28548622,0.15856248,-0.263661,0.03553435,0.09794897,-0.19086263,0.21170188,-0.035333,0.15629293,-0.20443127,0.059388347,-0.13384677,0.32383347,-0.01786511,0.026809739,0.09618991,0.16817331,0.003260169,0.1541657,0.14634871,-0.411643,-0.073803596,-0.05840593;-0.26815465,-0.07722028,0.4496276,0.21156041,0.34744376,0.14952148,-0.066632986,-0.02122758,-0.042498987,0.11149241,0.2645451,-0.2559394,0.09054535,-0.10177575,-0.09572797,0.13538697,-0.29660067,0.2413567,-0.09304269,0.16041638,0.052655313,0.066411205,-0.18594371,-0.20931654,0.14671618,0.12412398,-0.084041335,0.011399981,0.064732976,0.08294264,0.038126078,-0.092904456;0.043818083,0.08378806,0.2407548,-0.22100683,0.123821825,-0.26085117,-0.18233098,0.035554677,-0.17970791,-0.100569226,0.2404474,-0.26094705,-0.04062212,-0.106088966,-0.12377278,-0.06011476,0.06810586,0.06144369,-0.272036,0.23100908,0.03263779,0.21242264,0.13833047,-0.055457644,0.41442773,0.044346776,-0.08256207,0.25997287,-0.24246137,0.123531245,-0.0344091,0.19667232","modeluserbirthday":625852800000,"modelusercncertcode":"0907,0328","modeluserlabel10code":"","modeluserlabel11code":"","modeluserlabel12code":"","modeluserlabel13code":"","modeluserlabel14code":"","modeluserlabel15code":"","modeluserlabel16code":"","modeluserlabel17code":"","modeluserlabel1code":"","modeluserlabel2code":"100728","modeluserlabel3code":"100735","modeluserlabel4code":"","modeluserlabel5code":"","modeluserlabel6code":"","modeluserlabel7code":"100722","modeluserlabel8code":"","modeluserlabel9code":"","modeluserlat":23.419987,"modeluserlon":113.227236,"modeluserpreferenceradius":10,"modeluserweightedskillsfusionvectorv2":"0.19060262,-0.3171043,0.15806104,0.15741087,-0.03709896,-0.18762343,0.0051413607,-0.020985486,-0.1324314,-0.06700488,-0.05180177,-0.09148338,0.110959865,-0.23562765,-0.14517401,0.24540693,0.002885151,0.15705015,-0.28657448,0.3515736,0.33855277,0.14788917,0.13993894,-0.10371854,0.2964938,0.15936716,-0.22558168,-0.038045608,0.019879872,-0.12954699,-0.024876263,0.11587191","modeluserworkstarttime":1341072000000,"modelworkindustrynew":"143","personallabel":1,"platform":"1","sex":"1","short_apply_context_workfunc_jobworkfunc_map":"0703:2414\u001d2302:335\u001d2304:282\u001d0711:240\u001d2303:202\u001d0602:75\u001d0601:46\u001d2301:41\u001d3012:41\u001d0702:36\u001d0604:29\u001d4121:27\u001d0603:24\u001d1810:20\u001d0201:19\u001d4301:19\u001d3105:19\u001d2305:18\u001d0202:18\u001d0611:18\u001d3103:17\u001dA0K9:15\u001d3101:15\u001d0705:13\u001d4002:13\u001d4702:13\u001d3009:12\u001d4610:12\u001d2310:11\u001d0304:10\u001d0301:9\u001d4704:9\u001d4207:8\u001d0302:8\u001d0203:8\u001dA0L6:8\u001d2606:7\u001d2106:7\u001d1701:7\u001d0235:7\u001d3904:7\u001d3106:7\u001d3501:7\u001d3001:6\u001d3015:6\u001d2608:6\u001d2001:6\u001d0226:6\u001d3109:6\u001d2306:5\u001d0306:5\u001d8050:5\u001d0338:5\u001d4303:5\u001d3902:5\u001d4601:4\u001d8302:4\u001d1107:4\u001d3004:4\u001d0303:4\u001d0701:4\u001d0710:4\u001d1401:4\u001d0606:4\u001d0605:4\u001d4302:4\u001d3111:4\u001d3110:4\u001d3505:4\u001d0232:4\u001d4313:4\u001d3508:4\u001d3507:4\u001d4604:3\u001d8304:3\u001d0829:3\u001d8305:3\u001d8049:3\u001d0834:3\u001d3318:3\u001d0610:3\u001d1302:3\u001d0230:3\u001dA0LK:3\u001dA0LE:3\u001d0584:3\u001d3503:3\u001d4714:3\u001d3514:2\u001d8801:2\u001d8301:2\u001d4611:2\u001d6514:2\u001d0709:2\u001d4112:2\u001d3014:2\u001d0707:2\u001d8306:2\u001d8047:2\u001d8201:2","short_apply_user_job_companysize_map":"","short_apply_user_job_companytype_map":"","short_apply_user_job_degreefrom_map":"","short_apply_user_job_industrytype1_map":"","short_apply_user_job_jobsalarydown_map":"","short_apply_user_job_jobsalaryup_map":"","short_apply_user_job_workfunc_map":"","short_apply_user_job_workyear_map":"","short_apply_user_same_language_job_total_map":"","short_apply_user_total_map":"0","short_chat_context_workfunc_jobworkfunc_map":"0703:529\u001d2302:69\u001d2304:65\u001d0711:48\u001d2303:32\u001d0602:16\u001d3012:14\u001d2606:10\u001d0603:9\u001d3105:7\u001d3101:6\u001d0702:6\u001d1810:5\u001d2301:5\u001d3508:5\u001d4301:4\u001d4704:4\u001d4121:4\u001d3904:4\u001d3103:4\u001dA0K9:3\u001d5101:3\u001d0601:3\u001d0202:3\u001d0604:3\u001d4207:3\u001d1107:3\u001d0330:3\u001d4610:3\u001d2608:3\u001d0305:3\u001d4002:3\u001d3903:3\u001d0324:2\u001d0201:2\u001d3009:2\u001d2310:2\u001d6512:2\u001d4302:2\u001d0705:2\u001d4303:2\u001d4702:2\u001d8047:2\u001d3111:2\u001d3109:2\u001d3902:2\u001d0235:2\u001d6510:1\u001d0203:1\u001d2102:1\u001d4005:1\u001d4601:1\u001d4009:1\u001d4405:1\u001d0605:1\u001d0336:1\u001d1701:1\u001d1106:1\u001d5704:1\u001d0812:1\u001d7101:1\u001d4112:1\u001d2130:1\u001d0303:1\u001dA0LK:1\u001d0226:1\u001d2207:1\u001d1833:1\u001d0304:1\u001d3415:1\u001d0302:1\u001d1113:1\u001d0584:1\u001d4703:1\u001d0306:1\u001d8202:1\u001d3504:1\u001d3503:1\u001d3106:1\u001d0232:1\u001d4114:1\u001d3509:1\u001d4118:1\u001d3506:1","short_chat_user_job_companysize_map":"","short_chat_user_job_companytype_map":"","short_chat_user_job_degreefrom_map":"","short_chat_user_job_industrytype1_map":"","short_chat_user_job_jobsalarydown_map":"","short_chat_user_job_jobsalaryup_map":"","short_chat_user_job_workfunc_map":"","short_chat_user_job_workyear_map":"","short_chat_user_same_language_job_total_map":"","short_chat_user_total_map":"0","short_exposure_context_workfunc_jobworkfunc_map":"0703:235651\u001d2302:24595\u001d2304:21929\u001d0711:16560\u001d2303:12829\u001d0602:6885\u001d3012:3943\u001d3105:2750\u001d0603:2420\u001d0601:2233\u001d0702:2064\u001d0202:2026\u001d4121:1973\u001d2301:1961\u001d0604:1757\u001d3001:1727\u001d0705:1662\u001d3009:1571\u001d3103:1563\u001d1810:1486\u001d4301:1366\u001d4002:1082\u001dA0K9:1067\u001d0201:1059\u001d3106:937\u001d0611:922\u001d0302:920\u001d0701:883\u001d3508:878\u001d0304:870\u001d4303:851\u001d4704:843\u001d0226:795\u001d2310:771\u001d0203:726\u001d3004:695\u001d3101:681\u001d2305:653\u001d4001:644\u001d3904:585\u001d2606:581\u001d4302:525\u001d0235:525\u001d3501:508\u001dA0L6:486\u001d4610:485\u001d8302:477\u001d0305:470\u001d1701:470\u001d0709:466\u001d4702:432\u001d0232:427\u001d8049:423\u001d8050:415\u001d0301:399\u001d3111:395\u001d3507:388\u001d3903:386\u001d2001:385\u001d3109:383\u001d3102:380\u001d2306:362\u001d4006:344\u001d0626:338\u001d0303:337\u001d2610:334\u001d5101:330\u001d3902:325\u001d8047:311\u001d2608:302\u001d0710:300\u001d0338:293\u001d3110:291\u001d4007:286\u001d0584:282\u001d4207:280\u001d0606:279\u001d3514:276\u001d0336:276\u001d3505:276\u001d3015:275\u001d1401:269\u001d0605:255\u001d4009:250\u001d4601:246\u001d4714:245\u001d2106:231\u001d0629:228\u001d8305:223\u001d1107:212\u001d5114:211\u001d0306:207\u001d3506:207\u001d0208:206\u001d2002:200\u001d2004:197\u001d4112:193\u001d0405:192\u001d8312:191\u001dA0LE:189","short_exposure_user_job_companysize_map":"","short_exposure_user_job_companytype_map":"","short_exposure_user_job_degreefrom_map":"","short_exposure_user_job_industrytype1_map":"","short_exposure_user_job_jobsalarydown_map":"","short_exposure_user_job_jobsalaryup_map":"","short_exposure_user_job_workfunc_map":"","short_exposure_user_job_workyear_map":"","short_exposure_user_same_language_job_total_map":"","short_exposure_user_total_map":"0","short_look_context_workfunc_jobworkfunc_map":"0703:20629\u001d2302:2625\u001d2304:2246\u001d0711:1734\u001d2303:1457\u001d0602:577\u001d3012:409\u001d0702:268\u001d2301:259\u001d3105:239\u001d0603:232\u001d0601:213\u001d4121:194\u001d0604:182\u001d1810:180\u001d0202:170\u001d3009:146\u001d4301:140\u001d3103:137\u001d0201:129\u001d3001:125\u001dA0K9:121\u001d2305:111\u001d0705:107\u001d0304:105\u001d4002:104\u001d3101:95\u001d4610:94\u001d0302:94\u001d2606:90\u001d2310:88\u001d3106:87\u001d0611:80\u001d4704:74\u001d3904:73\u001d4303:71\u001d0235:64\u001d1701:62\u001d4702:56\u001d0203:55\u001d3508:55\u001d3004:52\u001d0301:51\u001d4302:48\u001dA0L6:47\u001d2306:45\u001d0303:45\u001d3111:45\u001d0226:44\u001d4207:43\u001d0305:43\u001d8302:42\u001d0701:42\u001d3109:41\u001d2608:40\u001d3902:39\u001d3903:38\u001d4006:37\u001d3015:36\u001d8050:36\u001d0605:35\u001d3501:35\u001d8049:33\u001d2001:31\u001d3102:31\u001d0710:29\u001d4007:29\u001d3110:29\u001d4601:28\u001d8047:28\u001d0232:28\u001d0709:27\u001d0584:27\u001d1401:26\u001d0606:26\u001d0610:26\u001d3905:26\u001d2106:25\u001d4001:24\u001d2610:24\u001d4009:24\u001d0336:24\u001d0626:24\u001d3505:24\u001d4313:24\u001d1107:23\u001d1302:22\u001d3507:22\u001d8305:21\u001d0829:20\u001d4314:20\u001d3506:20\u001d3514:19\u001d3014:19\u001d0208:19\u001d0338:19\u001d3503:19\u001d8312:18\u001d0337:17\u001dA0LK:17","short_look_user_job_companysize_map":"","short_look_user_job_companytype_map":"","short_look_user_job_degreefrom_map":"","short_look_user_job_industrytype1_map":"","short_look_user_job_jobsalarydown_map":"","short_look_user_job_jobsalaryup_map":"","short_look_user_job_workfunc_map":"","short_look_user_job_workyear_map":"","short_look_user_same_language_job_total_map":"","short_look_user_total_map":"0","topdegree":"6","topmajor":"0401","topschooltype":"0","workfunc":"0703","workindustry":"05"},"item_list":[{"item_id":"151576484"},{"item_id":"154118010"},{"item_id":"156179767"},{"item_id":"154419555"},{"item_id":"153630308"},{"item_id":"155851862"},{"item_id":"131512114"},{"item_id":"152093690"},{"item_id":"152942113"},{"item_id":"156125753"},{"item_id":"155849153"},{"item_id":"152793690"},{"item_id":"154641764"},{"item_id":"154832172"},{"item_id":"151683305"},{"item_id":"142495276"},{"item_id":"141901256"},{"item_id":"153054132"},{"item_id":"152414126"},{"item_id":"156024579"},{"item_id":"151440928"},{"item_id":"149528295"},{"item_id":"145015071"},{"item_id":"156224517"},{"item_id":"156105650"},{"item_id":"152727673"},{"item_id":"152207907"},{"item_id":"155826596"},{"item_id":"102904615"},{"item_id":"154260354"},{"item_id":"153966175"},{"item_id":"151524881"},{"item_id":"154250367"},{"item_id":"148613152"},{"item_id":"156179407"},{"item_id":"152176015"},{"item_id":"154499056"},{"item_id":"149724528"},{"item_id":"152855664"},{"item_id":"139841864"},{"item_id":"156024283"},{"item_id":"151735339"},{"item_id":"153357483"},{"item_id":"154306215"},{"item_id":"154051125"},{"item_id":"152365337"},{"item_id":"133739156"},{"item_id":"66058472"},{"item_id":"149157011"},{"item_id":"152205169"},{"item_id":"150803828"},{"item_id":"154816581"},{"item_id":"154367867"},{"item_id":"153327814"},{"item_id":"156051630"},{"item_id":"151105483"},{"item_id":"151636878"},{"item_id":"156049575"},{"item_id":"150587420"},{"item_id":"152577305"},{"item_id":"156091617"},{"item_id":"152598655"},{"item_id":"149845806"},{"item_id":"154739521"},{"item_id":"140980202"},{"item_id":"154549479"},{"item_id":"152117802"},{"item_id":"156056825"},{"item_id":"153425450"},{"item_id":"155508673"},{"item_id":"154105632"},{"item_id":"150189652"},{"item_id":"156131505"},{"item_id":"150118005"},{"item_id":"152174269"},{"item_id":"152550372"},{"item_id":"154662272"},{"item_id":"149010750"},{"item_id":"135464732"},{"item_id":"115931029"},{"item_id":"144071343"},{"item_id":"156066266"},{"item_id":"152222239"},{"item_id":"146248140"},{"item_id":"154161640"},{"item_id":"154188396"},{"item_id":"145944965"},{"item_id":"156112288"},{"item_id":"153864943"},{"item_id":"156105526"},{"item_id":"154701178"},{"item_id":"156212959"}],"request_id":"123","scene_id":"home_feed","size":92,"uid":"222539680"}
	`

func BenchmarkJsonUnmarshal(b *testing.B) {

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := make(map[string]any)
		if err := json.Unmarshal([]byte(data), &req); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJsonIterUnmarshal(b *testing.B) {

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := make(map[string]any)
		if err := json_iter.Unmarshal([]byte(data), &req); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJsonMarshal(b *testing.B) {
	req := make(map[string]any)
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}

func BenchmarkJsonIterMarshal(b *testing.B) {
	req := make(map[string]any)
	if err := json_iter.Unmarshal([]byte(data), &req); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json_iter.Marshal(req)
	}
}

var json_iter_fast = jsoniter.ConfigFastest

func BenchmarkJsonIterFastMarshal(b *testing.B) {
	req := make(map[string]any)
	if err := json_iter.Unmarshal([]byte(data), &req); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json_iter_fast.Marshal(req)
	}
}
