package validator

import (
	"fmt"
	"ginblog/utils/errmsg"
	"github.com/go-playground/locales/zh_Hans_CN"
	unTrans "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
)

// Go 1.18引入泛型，any就是interface{}的别名
// 可以传进去接口，一般都是结构体进行验证，统一函数
func Validate(data any) (string, int) {
	//实例化
	validate := validator.New()
	//要转换为中文
	uni := unTrans.New(zh_Hans_CN.New())
	//翻译方法
	trans, _ := uni.GetTranslator("zh_Hans_CN")
	//翻译方法注册
	err := zhTrans.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		fmt.Println("err:", err)
	}
	//标签反射，不然username翻译过来还是username
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		label := field.Tag.Get("label")
		return label
	})

	//保证了输入的data一定是一个结构体
	err = validate.Struct(data)
	if err != nil {
		//.() 语法是类型断言
		for _, v := range err.(validator.ValidationErrors) {
			return v.Translate(trans), errmsg.ERROR
		}
	}
	return "", errmsg.SUCCES
}
