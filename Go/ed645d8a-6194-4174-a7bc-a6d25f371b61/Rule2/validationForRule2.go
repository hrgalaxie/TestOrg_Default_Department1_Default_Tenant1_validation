
	package validation
	import (
		"bufio"
		"fmt"
		"github.com/antlr/antlr4/runtime/Go/antlr"
		// "golang.org/x/sync/syncmap"
		"sync"
		"os"
		"sort"
		"strconv"
		"strings"
		parser "TestOrg_Default_Department1_Default_Tenant1_validation/Go/ed645d8a-6194-4174-a7bc-a6d25f371b61/Rule2/parser"
		

	)

	
	//var everyRule[] string
	var ErrorListMap sync.Map
	var ErrorHandleMap sync.Map
	var RuleArray []string
	var SizeArray []string
	type pluginFunction string
	type ErrorInfo struct{
		RuleName  string
		ErrorDetails []ErrorDetails
	}																		

	type ErrorDetails struct{
		FieldName string
		Value interface{}
		Error     string
	}
	
	func (g pluginFunction) CallValidation(input map[string]string,runID string) (bool,interface{}){
		//everyRule = nil
		ok,errorInfo := Validate(input,runID)
		fmt.Println("RuleName:",errorInfo.RuleName)
		return ok,errorInfo
	}
	var ValidationPlugin pluginFunction

	type OrderedMap struct {
		Order []string
		Map map[string]map[string]interface{}
	}


	type ValidationListener struct {
		*parser.BaseRule2Listener
	}

	type ErrorListener struct {
		antlr.DefaultErrorListener
		RunID string
	}
	func (l ValidationListener) EnterEveryRule(ctx antlr.ParserRuleContext){
		//everyRule=append(everyRule,ctx.GetRuleContext().GetText())
	}

	func (l ErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException)  {
		ErrorHandleMap.Store(l.RunID,true)
		errorFound:=strconv.Itoa(column)+"-->"+msg
		if outerMap, ok := ErrorListMap.Load(l.RunID); ok {
			if value, canConvert := outerMap.([]string); canConvert {
				fmt.Println("Error:::",value)
				value = append(value,errorFound)
				ErrorListMap.Store(l.RunID,value)
			}
		}else{
			value := []string{errorFound}
			ErrorListMap.Store(l.RunID,value)
		}
		fmt.Fprintln(os.Stderr, "line "+strconv.Itoa(line)+":"+strconv.Itoa(column)+" "+msg)
	}

	

	func getGrammarFile() string{

		file := `grammar Rule2;

expression :  k n p ph rainfall temperature EOF ;

k : K ;

n : N ;

p : P ;

ph : PH ;

rainfall : RAINFALL ;

temperature : TEMPERATURE ;

K :SPLITKEY  OWNKEY [0-9]* ;

N :SPLITKEY  OWNKEY OWNKEY [0-9]* ;

P :SPLITKEY  OWNKEY OWNKEY OWNKEY [0-9]* ;

PH :SPLITKEY  OWNKEY OWNKEY OWNKEY OWNKEY [0-9]* ;

RAINFALL :SPLITKEY  OWNKEY OWNKEY OWNKEY OWNKEY OWNKEY [0-9]* ;

TEMPERATURE :SPLITKEY  OWNKEY OWNKEY OWNKEY OWNKEY OWNKEY OWNKEY [0-9]* ;

OWNKEY : '$~$' ;

SPLITKEY : ' !$~$#%#&$&! ' ;

WS : [ \t\n\r]+ -> skip ;

/*
LimitBegin
1 0,* length
2 0,* length
3 0,* length
4 0,* length
5 0,* length
6 0,* length
End
*/
`
		
		return file

    }


	func ReadFromFile() ([]string,error){
		
		file := getGrammarFile()		

		var start bool

		var sizeArray[] string

		start=false
		
		scanner := bufio.NewScanner(strings.NewReader(file))
		
		for scanner.Scan(){
			temp:=scanner.Text()

			if temp=="LimitBegin"{
				start=true
				continue
			}
			if temp=="End"{
				break
			}
			if start{
				//fmt.Println(temp)
				sizeArray=append(sizeArray,temp)
			}
		}

		if err:=scanner.Err();err!=nil{
			//log.Fatal(err)
			fmt.Println(err)
			return sizeArray,err
		}

		return sizeArray,nil

	}

	func KeySortAccordingToGrammar(key []string) []string{

		if RuleArray == nil || len(RuleArray) == 0 {
			RuleArray = []string{}
			file := getGrammarFile()
			temp := ""
			scanner := bufio.NewScanner(strings.NewReader(file))
			for scanner.Scan(){
				temp = scanner.Text()

				if strings.Contains(temp,"expression"){
					break
				}
			}

			splitArray := strings.Split(temp,":")
			if len(splitArray)==2{
				temp = splitArray[1];
				temp = strings.TrimSpace(temp)
				split := strings.Split(temp," ")

				for _,value := range split{

					if value != "EOF" && value != ";" {
						RuleArray = append(RuleArray,value)
					}

				}

			}


		}
		
		var emptyArray []string

		if len(RuleArray) != len(key){
			return emptyArray
		}

		for _,first := range RuleArray{
			found := false
			for _,second := range key{
				if first == second{
					found = true
					break
				}
			}
			if !found{
				return emptyArray;
			}
		}

		return RuleArray;


	}


	
	func Validate(input map[string]string,runID string) (bool,ErrorInfo){

		//Ownkey
		ownKey:="$~$"

		var keys []string

		for k,_ := range input{
			keys = append(keys,k)
		}

		keySortToGrammar := KeySortAccordingToGrammar(keys)

		if len(keySortToGrammar)==0{
			if len(keys) == 0{
				return true,ErrorInfo{}
			}else{
				return false,ErrorInfo{}
			}
			
		}

		inputTemp:=""
		var lineHashMap = make(map[int]string)
		for i,value :=range keySortToGrammar{

			temp:=""

			for j:=0;j<i+1;j++{

				temp=ownKey+temp

			}

			inputTemp = inputTemp+" !$~$#%#&$&! "+temp+input[value]
			fmt.Println("len of inputTemp:",len(inputTemp))
			lineHashMap[len(inputTemp)] = value
		}
		ok,errorInfo := ValidateInput(inputTemp,runID,lineHashMap)
		if ok{
			fmt.Println("Corrrect data")
			return true,ErrorInfo{}
		}

		fmt.Println("Incorrect data")
		return false,errorInfo
	}
	func toString(array []string) string{
		temp:=""
		for _,st := range array{
			st=strings.Trim(st,"$~$")
			temp=temp+st+"\n"
		}
		return temp
	}

	func ValidateInput(input string,runID string,lineHashMap map[int]string) (bool,ErrorInfo){
		ErrorHandleMap.Store(runID, false)
		fmt.Println("Input ===>",input)
		in:=antlr.NewInputStream(input)
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error1:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false, ErrorInfo{}
							}
						}
					}

				}
			}
		}
		lexer:=parser.NewRule2Lexer(in)
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error2:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		lexer.RemoveErrorListeners()
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error3:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		lexer.AddErrorListener(&ErrorListener{antlr.DefaultErrorListener{},runID})
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error4:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		stream:=antlr.NewCommonTokenStream(lexer,0)
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error5:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		parser:=parser.NewRule2Parser(stream)
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error6:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		parser.RemoveErrorListeners()
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error7:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		parser.AddErrorListener(&ErrorListener{antlr.DefaultErrorListener{},runID})
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error8:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		parser.BuildParseTrees=true

		tree:=parser.Expression()
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error9:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							fmt.Println("Length Of Error List:",len(errorList))
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								fmt.Println("Error # ", result.RuleName)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		antlr.ParseTreeWalkerDefault.Walk(&ValidationListener{},tree)
		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error10:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}
			}
		}
		inputList := strings.Split(input," !$~$#%#&$&! ")
		fmt.Println("iNput inputList",inputList)
		fmt.Println(len(inputList))
		if SizeArray == nil || len(SizeArray) == 0{
			SizeArray = []string{}
			sizeArray,scannerErr:=ReadFromFile()
			if scannerErr != nil{
				return false,ErrorInfo{}
			}
			SizeArray = sizeArray
		}
		
		if len(SizeArray)>=1{
				for _,v := range SizeArray{
					temp:=strings.Split(v," ")

					index,er1:=strconv.Atoi(temp[0])
					if er1!=nil{
						fmt.Println(er1)
						return false,ErrorInfo{}
					}
					hashMap := make(map[string]string)
					for index,value := range temp{

						if value == "limit"{
							hashMap[value] = temp[index-1]
						}

						if value == "length"{
							hashMap[value] = temp[index-1]
						}

						if value == "include"{
							hashMap[value] = temp[index-1]
						}

						if value == "exclude"{
							hashMap[value] = temp[index-1]
						}
					}

					//fmt.Println(everyRule,"every rule")
					ruleValue:=inputList[index]
					if rangeValue,ok := hashMap["length"]; ok {
						minmax := strings.Split(rangeValue,",")
						minStr := ""
						maxStr := ""
						infiniteMaxValue := ""
						if len(minmax) > 0{
							minStr = minmax[0]
							if minmax[1] == "*"{
								infiniteMaxValue = minmax[1]
								maxStr = "0.0"
							}else{
								maxStr = minmax[1]
							}
						}
						min,er2:=strconv.ParseFloat(minStr,64)
						max,err3 := strconv.ParseFloat(maxStr,64)
						if er2!=nil{
							fmt.Println(er2)
							return false,ErrorInfo{}
						}
						if err3!=nil{
							fmt.Println(err3)
							return false,ErrorInfo{}
						}

						tempIndex:=index
						tempIndex*=3
						if infiniteMaxValue == "*"{
							if len(ruleValue)-tempIndex >= int(min){

							}else {
								fmt.Fprintln(os.Stderr,ruleValue+"length is less than defined minimun length")
								result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"length is less than defined minimun length")
								return false,result
							}
						}else{
							if len(ruleValue)-tempIndex >= int(min) && len(ruleValue)-tempIndex <= int(max){

							}else {
								fmt.Fprintln(os.Stderr,ruleValue+"exceed length")
								result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"exceed length")
								return false,result
							}
						}

					}
					if  excludeValue,ok:= hashMap["exclude"];ok{
						tempV,e:=strconv.ParseFloat(strings.Trim(ruleValue,"$~$"),64)
						if e!=nil{
							fmt.Println(e)
							return false,ErrorInfo{}
						}

						excludeArr := strings.Split(excludeValue,",")
						for _,exValue := range excludeArr{
							exclude,err := strconv.ParseFloat(exValue,64)
							if err!=nil{
								fmt.Println(err)
								return false,ErrorInfo{}
							}
							if tempV == exclude{
								fmt.Fprintln(os.Stderr,ruleValue+"value is including in the exclude value list")
								result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"value is including in the exclude value list")
								return false,result
							}
						}

					}
					if rangeValue,ok := hashMap["limit"]; ok {
						fmt.Println("rangeValue:" , rangeValue)
						minmax := strings.Split(rangeValue,",")
						minStr := ""
						maxStr := ""
						infiniteMaxValue := ""
						fmt.Println("Len of Min Max:",len(minmax))
						if len(minmax) > 0{
							minStr = minmax[0]
							if minmax[1] == "*"{
								infiniteMaxValue = minmax[1]
								maxStr = "0.0"
							}else{
								maxStr = minmax[1]
							}
						}
						min,er2:=strconv.ParseFloat(minStr,64)
						max,err3 := strconv.ParseFloat(maxStr,64)
						if er2!=nil{
							fmt.Println(er2)
							return false,ErrorInfo{}
						}
						if err3!=nil{
							fmt.Println(err3)
							return false,ErrorInfo{}
						}

						fmt.Println("it is limit")

						tempV,e:=strconv.ParseFloat(strings.Trim(ruleValue,"$~$"),64)
						if e!=nil{
							fmt.Println(e)
							return false,ErrorInfo{}
						}
						if infiniteMaxValue == "*"{
							if tempV>=min{

							}else{

								if includeValue,ok:= hashMap["include"];ok{
									includeArr := strings.Split(includeValue,",")
									isInclude := false
									for _,include := range includeArr{
										include,err := strconv.ParseFloat(include,64)
										if err!=nil{
											fmt.Println(err)
											return false,ErrorInfo{}
										}
										if tempV == include{
											isInclude = true
											break
										}
									}
									if !isInclude{
										fmt.Fprintln(os.Stderr,ruleValue+" value is less than minimun limit value")
										result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"value is less than minimun limit value")
								        return false,result
										
									}
									

								}else{

									fmt.Fprintln(os.Stderr,ruleValue+" value is less than minimun limit value")
									result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"value is less than minimun limit value")
								    return false,result
								}

							}
						}else{
							if tempV>=min && tempV <= max{

							}else{

								if includeValue,ok:= hashMap["include"];ok{
									includeArr := strings.Split(includeValue,",")
									isInclude := false
									for _,include := range includeArr{
										include,err := strconv.ParseFloat(include,64)
										if err!=nil{
											fmt.Println(err)
											return false,ErrorInfo{}
										}
										if tempV == include{
											isInclude = true
											break
										}
									}
									if !isInclude{

										fmt.Fprintln(os.Stderr,ruleValue+"exceed limit value")
									    result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"exceed limit value")
								        return false,result
									}

								}else{

									fmt.Fprintln(os.Stderr,ruleValue+"exceed limit value")
									result := HandleCustomError(lineHashMap,index,strings.Trim(ruleValue,"$~$"),"exceed limit value")
								    return false,result
								}


							}
						}

					}


				}
			}


		if outerMap, ok := ErrorHandleMap.Load(runID); ok {
			if value, canConvert := outerMap.(bool); canConvert {
				fmt.Println("Error1:",value)
				if value{
					fmt.Println("Error exist")
					if errorMap, Ok := ErrorListMap.Load(runID); Ok {
						if errorList,isConvert := errorMap.([]string);isConvert{
							if len(errorList) != 0{
								result := handleError(errorList,input,lineHashMap)
								return false,result
							}else{
								return false,ErrorInfo{}
							}
						}
					}

				}else{
					fmt.Println("No Error exist")
					return true,ErrorInfo{}
				}
			}
		}

		return true,ErrorInfo{}
	}

	func HandleCustomError(lineHashMap map[int]string,index int,value string,errorMessage string)ErrorInfo{
		var errorInfo ErrorInfo
		errorInfo.RuleName = "Rule2"
		var errorDetails []ErrorDetails
		var errorDetail ErrorDetails
		var positionList []int
		for k,_ := range lineHashMap{

			positionList = append(positionList,k)
				
		}
		sort.Ints(positionList)
		
		fieldName := lineHashMap[positionList[index-1]]
		errorDetail.FieldName = fieldName
		errorDetail.Value = value
		errorDetail.Error = errorMessage
		errorDetails = append(errorDetails,errorDetail)
		fmt.Println("RuleName :::",errorInfo.RuleName)
		errorInfo.ErrorDetails = errorDetails
		return errorInfo


	}


	func handleError(errorList []string,input string,lineHashMap map[int]string) ErrorInfo{

		var errorInfo ErrorInfo
		errorInfo.RuleName = "Rule2"

		//inputLength := len(input)
		fmt.Println("input ***>",input)
		inputList := strings.Split(input," !$~$#%#&$&! ")

		input = strings.Replace(input," !$~$#%#&$&! ","",-1)
		var errorDetails []ErrorDetails
		for _,err := range errorList{
			var errorDetail ErrorDetails
			tempSplit := strings.Split(err,"-->")

			if len(tempSplit)!=2{
				continue
			}

			index,err := strconv.Atoi(tempSplit[0])
			fmt.Println("index -->", index)
			var positionList []int
			for k,_ := range lineHashMap{

				positionList = append(positionList,k)
				
			}
			sort.Ints(positionList)
			var fieldName string
			for _,position := range positionList{
				fmt.Println("index::",position)
				if index < position {
					fieldName = lineHashMap[position]
					break
				}
			}
			errorDetail.FieldName = fieldName
			errorString := tempSplit[1]
			if err != nil{
				continue
			}

			multiplier := 3

			end := 0

			for i,value := range inputList{
				value = strings.Replace(value,"$~$","",-1)
				if value == ""{
					continue
				}

				fmt.Println("value ====> ", value)
				start := end + 1
				end = multiplier + start +  13  +len(value) 
				multiplier = multiplier * (i+1)

				if index>= start && index <= end{
					errorString = strings.Replace(errorString," '$~$'","",-1 )
					temp := "FieldName:" + fieldName +",Value:" +value+","+errorString
					fmt.Println("Error ===> ",temp)
					errorDetail.Value = value
					errorDetail.Error = errorString
					errorDetails = append(errorDetails,errorDetail)
				}
			}
		}
		fmt.Println("RuleName :::",errorInfo.RuleName)
		errorInfo.ErrorDetails = errorDetails
		return errorInfo
	}

	
