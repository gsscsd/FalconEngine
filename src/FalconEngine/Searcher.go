/*****************************************************************************
 *  file name : Searcher.go
 *  author : Wu Yinghao
 *  email  : wyh817@gmail.com
 *
 *  file description : 搜索引擎
 *
******************************************************************************/


package main

import (
	"BaseFunctions"
	"indexer"
	"strconv"
	"net/url"
	"encoding/json"
	"utils"

)



type Searcher struct{
	*BaseFunctions.BaseProcessor
	Indexer		*indexer.IndexSet
}


func NewSearcher(processor *BaseFunctions.BaseProcessor,indexer *indexer.IndexSet) *Searcher{
	this:=&Searcher{processor,indexer}
	return this
}

const PAGE_NUM string = "pg"
const PAGE_SIZE string = "ps"
const SORT_BY	string = "sort_by"
const GROUP_BY	string = "group_by"
const QUERY		string = "query"


func (this *Searcher)Process(log_id string,body []byte,params map[string]string , result map[string]interface{},ftime func(string)string) error {
	
	
	_,has_ctl := params["_contrl"]
	if has_ctl{
		this.Indexer.GetIndexInfo(result)
		return nil
	}
	
	this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("Process running"))
	searchRules,err :=this.ParseSearchInfo(log_id,params,body)
	if err !=nil {
		this.Logger.Error("[LOG_ID:%v]Running Searcher ..err : %v ..Time: %v ",log_id,err,ftime("search fields"))
	}

	//srules,frules,_,_ := this.ParseParams(log_id,params)
	this.Logger.Info("[LOG_ID:%v]Running Searcher  %v....Time: %v ",log_id,searchRules,ftime("ParseSearchInfo"))
	
	total_doc_ids:=make([]utils.DocIdInfo,0)
	for _,search_rule := range searchRules{
		
		doc_ids,ok:=this.Indexer.SearchByRules(search_rule.SR)
		if !ok{
			result["DATA"]="NO DATA"
			//return nil
		}	
		this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("search fields"))
		doc_ids,_ = this.Indexer.FilterByRules(doc_ids,search_rule.FR)
		this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("fliter fields"))
		//this.Logger.Info("[LOG_ID:%v]Running Searcher ....RES::: %v ",log_id,doc_ids)
		//this.Logger.Info("[LOG_ID:%v]Running Searcher ....doc_ids::: %v ",log_id,doc_ids)
		total_doc_ids,_ = utils.Merge(total_doc_ids,doc_ids) 
		//this.Logger.Info("[LOG_ID:%v]Running Searcher ....total_doc_ids::: %v ",log_id,total_doc_ids)
	}
	//this.Logger.Info("[LOG_ID:%v]Running Searcher ....RES::: %v ",log_id,total_doc_ids)
	var tmp_doc_ids  []utils.DocIdInfo
	if len(total_doc_ids) > 20{
		tmp_doc_ids = total_doc_ids[:15]
	}else{
		tmp_doc_ids = total_doc_ids
	}
	ids,fields:=this.Indexer.GetDetails(tmp_doc_ids)
	var infos []map[string]string
	for _,id := range ids {
		info,err:=this.RedisCli.GetFields(id,fields)
		if err != nil {
			this.Logger.Error("%v",err)
		}
		infos=append(infos,info)
	}
	this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("Display Detail"))
	result["DATA"]=infos
	/*
	srules,frules,_,_ := this.ParseParams(log_id,params)

	doc_ids,ok:=this.Indexer.SearchByRules(srules)
	if !ok{
		result["DATA"]="NO DATA"
		return nil
	}	
	this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("search fields"))
	doc_ids,_ = this.Indexer.FilterByRules(doc_ids,frules)
	this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("fliter fields"))
	
	

	ids,fields:=this.Indexer.GetDetails(doc_ids)
	var infos []map[string]string
	for _,id := range ids {
		info,err:=this.RedisCli.GetFields(id,fields)
		if err != nil {
			this.Logger.Error("%v",err)
		}
		infos=append(infos,info)
	}
	this.Logger.Info("[LOG_ID:%v]Running Searcher ....Time: %v ",log_id,ftime("Display Detail"))
	result["DATA"]=infos
	//
	//result["PAGES"] = len(doc_ids)/int(ps) + 1

	//result["DATA"]=doc_ids
	*/
	
	
	return nil
}


type Condition struct {
	Key		string `json:"key"`
	Op 		string `json:"operate"`
	Value	string `json:"value"`
	Desc    string `json:"desc"`
	
}

type ConditionData struct {
	
	Data  [][]Condition  `json:"data"`
}


type SearchInfo struct {
    Customer_id      int64       `json:"customer_id"`
    Id               int64       `json:"_id"`
    Creator_id       int64       `json:"creator_id"`
    Last_editor_id   int64       `json:"last_editor_id"`
    Create_time      string      `json:"create_time"`
    Last_modify_time string      `json:"last_modify_time"`
    Editor_id        int64       `json:"editor_id"`
    Group_type       int64       `json:"group_type"`
    Name             string      `json:"name"`
    Conditions       ConditionData `json:"conditions"`

}


type SearchRules struct {
	SR []indexer.SearchRule
	FR []indexer.FilterRule
}


func (this *Searcher) ParseSearchInfo(log_id string,params map[string]string,body []byte) ([]SearchRules,error) {// ([]indexer.SearchRule,[]indexer.FilterRule,int64,int64){

	var searchInfo SearchInfo

	err := json.Unmarshal(body, &searchInfo)
	if err != nil {
		this.Logger.Error("[LOG_ID:%v]  %v", log_id, err)
		return nil,err
	}


	searchrules := make([]SearchRules,0)
	
	this.Logger.Info("SearchInfo : %v \n",searchInfo)
	for i,v := range searchInfo.Conditions.Data{
		this.Logger.Info("Conditions[%v] : %v \n",i,v)
		var SRs SearchRules
		SRs.SR = make([]indexer.SearchRule,0)
		SRs.FR = make([]indexer.FilterRule,0)
		SRs.SR = append(SRs.SR,indexer.SearchRule{Field:"cid",Query:searchInfo.Customer_id})
		for ii,vv := range v {
			this.Logger.Info("\t\t Condition[%v] : %v \n",ii,vv)
			if vv.Key == "user_attrib" {
				//如果是包含，表示倒排检索
				if vv.Op == "include" {
					var SR indexer.SearchRule
					SR.Field = vv.Desc
					SR.Query = vv.Value
					SRs.SR=append(SRs.SR,SR)
				}else{ //正排检索
					var FR indexer.FilterRule
					FR.Field = vv.Desc
					FR.Value = vv.Value
					FR.IsForward = true 
					switch vv.Op{
						case "less":
							FR.FiltType = indexer.FILT_TYPE_LESS
						case "more":
							FR.FiltType = indexer.FILT_TYPE_ABOVE
						case "equal":
							FR.FiltType = indexer.FILT_TYPE_EQUAL
						case "unequal":
							FR.FiltType = indexer.FILT_TYPE_UNEQUAL
						default:
							FR.FiltType = indexer.FILT_TYPE_LESS
					}
					SRs.FR = append(SRs.FR,FR)
				}
			}else{
				
				if vv.Key == "mail"{  //如果是邮件，需要拼接字符串，特殊处理
					var SR indexer.SearchRule
					switch vv.Op{
						case "look": //查看
							SR.Query = vv.Value + "_" + "1"
							SR.Field = "email_view"
						case "click":
							SR.Query = vv.Value + "_" + "1"
							SR.Field = "email_click"
						case "send":
							SR.Query = vv.Value + "_" + "1"
							SR.Field = "email_sended"
						case "unlook":
							SR.Query = vv.Value + "_" + "0"
							SR.Field = "email_view"
						case "unclick":
							SR.Query = vv.Value + "_" + "0"
							SR.Field = "email_click"
						case "unsend":
							SR.Query = vv.Value + "_" + "0"
							SR.Field = "email_sended"
						default:
						
					}
					SRs.SR=append(SRs.SR,SR)
					continue
					
				}
				
				if vv.Key == "sms" {//如果是短信，需要拼接字符串,特殊处理
					var SR indexer.SearchRule
					switch vv.Op{
						case "click":
							SR.Query = vv.Value + "_" + "1"
							SR.Field = "sms_click"
						case "send":
							SR.Query = vv.Value + "_" + "1"
							SR.Field = "sms_sended"
						case "unclick":
							SR.Query = vv.Value + "_" + "0"
							SR.Field = "sms_click"
						case "unsend":
							SR.Query = vv.Value + "_" + "0"
							SR.Field = "sms_sended"
						default:

					}
					SRs.SR=append(SRs.SR,SR)
					continue
					
				}
				
				if vv.Key == "area"{ //如果是地区，需要特殊处理 TODO
					from_num , err := strconv.ParseInt(vv.Value, 0, 0)
					if err != nil {
						continue
					}
					//from_source
					var FR indexer.FilterRule
					FR.Field = "from_source"
					FR.Value = from_num
					FR.IsForward = true
					
					
					if from_num < 1000{
						//FR.Value = from_num*10000
						//FR.FiltType = indexer.FILT_TYPE_EQUAL
						//SRs.FR = append(SRs.FR,FR)
						
						FR.FiltType = indexer.FILT_TYPE_ABOVE
						SRs.FR = append(SRs.FR,FR)
						
						
						FR.Value = from_num*10000 + 10000
						FR.FiltType = indexer.FILT_TYPE_LESS
						SRs.FR = append(SRs.FR,FR)
						continue
					}
					
					if from_num < 100000{
						FR.FiltType = indexer.FILT_TYPE_ABOVE
						SRs.FR = append(SRs.FR,FR)
						FR.FiltType = indexer.FILT_TYPE_LESS
						FR.Value = from_num+100
						SRs.FR = append(SRs.FR,FR)
						continue
					}
					
					continue
				}
				
				if vv.Key == "source"{
					var FR indexer.FilterRule
					FR.Field = vv.Key
					FR.IsForward = true
					var addbyadmin int64 = 1 
					var export int64 = 2
					var unknown int64 = 3 
					switch vv.Value{
						case "addbyadmin":
							FR.Value = addbyadmin
						case "export":
							FR.Value = export						
						case "unknown":
							FR.Value = unknown					
						default:
							FR.Value = unknown
					}
					if vv.Op == "equal"{
							FR.FiltType = indexer.FILT_TYPE_EQUAL	
					}else{
							FR.FiltType = indexer.FILT_TYPE_UNEQUAL	
					}
					SRs.FR = append(SRs.FR,FR)
					continue 
				}
				
				
				if vv.Key == "email_client"{  //TODO
					
				}
				
				if vv.Op == "include" { //其他检索，倒排索引
				
					var SR indexer.SearchRule
					SR.Field = vv.Desc
					SR.Query = vv.Value
					SRs.SR=append(SRs.SR,SR)
					
				}else{  //其他检索，正排索引
					
					var FR indexer.FilterRule
					FR.Field = vv.Key
					FR.Value = vv.Value
					FR.IsForward = true 
					switch vv.Op{
						case "less":
							FR.FiltType = indexer.FILT_TYPE_LESS
						case "more":
							FR.FiltType = indexer.FILT_TYPE_ABOVE
						case "equal":
							FR.FiltType = indexer.FILT_TYPE_EQUAL
						case "unequal":
							FR.FiltType = indexer.FILT_TYPE_UNEQUAL
						default:
							FR.FiltType = indexer.FILT_TYPE_LESS
					}
					SRs.FR = append(SRs.FR,FR)
				}
				
			
				
			}
			
		} // end for ii,vv := range v
		searchrules = append(searchrules,SRs)
	}

	return searchrules,nil
	
}


func (this *Searcher) ParseParams(log_id string,params map[string]string) ([]indexer.SearchRule,[]indexer.FilterRule,int64,int64){
	
	srules:=make([]indexer.SearchRule,0)
	frules:=make([]indexer.FilterRule,0)
	
	
	var ps int64
	var pg int64
	var err error
	
	ps=10
	pg=1
	for k,v := range params{
		v, _ = url.QueryUnescape(v)
		if k == PAGE_NUM{
			pg, err = strconv.ParseInt(params[PAGE_NUM], 0, 0)
			if err != nil {
				ps=10
			}
			continue
		}
		
		if k == PAGE_SIZE{
			ps, err = strconv.ParseInt(params[PAGE_SIZE], 0, 0)
			if err != nil {
				pg=1
			}
			continue
		}
		
		if k == QUERY {
			//this.Logger.Info(" query K : %v ,V : %v",k,v)
			srules=append(srules,indexer.SearchRule{k,v})
			continue
		}
		
		
		if k[0] != '-' && k[0] != '_' {
			this.Logger.Info(" string field K : %v ,V : %v",k,v)
			stype := this.Indexer.GetIdxType(k)
			if stype == -1 {
				continue
			}
			if stype ==1 {
				srules=append(srules,indexer.SearchRule{k,v})
			}else{
				v_n, err := strconv.ParseInt(v, 0, 0)
				if err != nil {
					this.Logger.Error("[LOG_ID:%v] %v %v", log_id, v,err)
					continue
				}
				srules=append(srules,indexer.SearchRule{k,v_n})
			}
			
			continue 
		}
		
		if k[0] == '-'{
			this.Logger.Info(" filter1 field K : %v ,V : %v",k,v)
			stype := this.Indexer.GetPflType(k[1:])
			if stype == -1 {
				this.Logger.Error("[LOG_ID:%v] %v %v", log_id, v,k[1:])
				continue
			}
			if stype ==1 {
				frules=append(frules,indexer.FilterRule{k[1:],true,1,v})
			}else{
				v_n, err := strconv.ParseInt(v, 0, 0)
				if err != nil {
					this.Logger.Error("[LOG_ID:%v] %v %v", log_id, v,err)
					continue
				}
				frules=append(frules,indexer.FilterRule{k[1:],true,1,v_n})
			}
			
			continue
		}
		
		if k[0] == '_'{
			this.Logger.Info(" filter2 field K : %v ,V : %v",k,v)
			stype := this.Indexer.GetPflType(k[1:])
			if stype == -1 {
				this.Logger.Error("[LOG_ID:%v] %v %v", log_id, v,k[1:])
				continue
			}
			if stype ==1 {
				frules=append(frules,indexer.FilterRule{k[1:],false,1,v})
			}else{
				v_n, err := strconv.ParseInt(v, 0, 0)
				if err != nil {
					this.Logger.Error("[LOG_ID:%v] %v %v", log_id, v_n,err)
					continue
				}
				frules=append(frules,indexer.FilterRule{k[1:],false,1,v_n})
			}
			continue
		}
		
	}
	
	
	return srules,frules,ps,pg
}




