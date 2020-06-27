
package parser

import (
	"fmt"
	"sort"
)

type ParseObject struct {
	ObjectType string
	Name string
	Index int
	Parent *ParseObject
	Values map[string]ParseValue
	Definitions map[string]*ParseObjectArray
}

func (obj *ParseObject) LookupValue(name string) ParseValue {
	v := obj.Values[name]
	if v != nil {
		return v
	}

	d := obj.Definitions[name]
	if d == nil {
		fmt.Printf("FIXME no value found for %v\n", name)
		fmt.Println(obj.Definitions)
		return nil
	}

	if d.Template.IndexNames[0] == "" { // FIXME maybe clarify with an IsArray method
		return ObjectValue { &d.Children[0] }
	} else {
		return ArrayValue { d }
	}
}

type ParseObjectArray struct {
	Children []ParseObject
	Template ObjectArrayTemplate
	Parent *ParseObject
}

type ObjectArrayTemplate struct {
	ObjectType string
	IndexSizes []int
	IndexNames []string
}

func (tp *ObjectArrayTemplate) CreateObjects(parser *Parser, object *ParseObject, sym Token) {

	// FIXME check that index symbols aren't duplicates or in the symbol table
	n := 1
	for _,x := range tp.IndexSizes {
		n = n*x
	}

	array := ParseObjectArray{}
	array.Template = *tp
	array.Parent = object
	array.Children = make([]ParseObject, n)
	for i:=0; i<n; i++ {
		array.Children[i].ObjectType = tp.ObjectType
		array.Children[i].Values = map[string]ParseValue{}
		array.Children[i].Definitions = map[string]*ParseObjectArray{}
		array.Children[i].Name = sym.Text
		array.Children[i].Index = i
		array.Children[i].Parent = object

	}
	object.Definitions[sym.Text] = &array
}

func (obj *ParseObject) Print(depth int) {

	keys := make([]string, len(obj.Values))
	i := 0
	for k,_ := range obj.Values {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _,key := range keys {
		for i=0; i<depth; i++ {
			fmt.Printf("\t")
		}
		fmt.Printf("%v: %v\n", key, obj.Values[key].ToString())
	}

	keys = make([]string, len(obj.Definitions))
	i = 0
	for k,_ := range obj.Definitions {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _,key := range keys {
		array := obj.Definitions[key]
		for _,child := range array.Children {
			for i=0; i<depth; i++ {
				fmt.Printf("\t")
			}
			fmt.Printf("%v %v#%v {\n", child.ObjectType, child.Name, child.Index)
			child.Print(depth+1)
			for i=0; i<depth; i++ {
				fmt.Printf("\t")
			}
			fmt.Printf("}\n")
		}
	}

}

