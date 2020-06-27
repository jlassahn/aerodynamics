
package parser

/*

tokens:
 whitespace is spaces, tabs, linesfeeds, comments
 comment is # ... EOL
 Alpha followed by any alphanumeric or _
 digit followed by digits and .
 any single nonalphanumeric


Grammar

File:
	*Assignment  # one or more

Assignment:
	Value
	Definition

Value:
	TOKEN : Expression

Definition:
	DefType DefName  { *Assignment }   # zero or more

DefType:
	TOKEN  # Tube, Sheet, Cap, Mount, ...

DefName:
	Name
	Name [ DefIndexList ]

Name:
	TOKEN

DefIndexList:
	DefIndex
	DefIndex , DefIndexList

DefIndex:
	TOKEN ~ Expression

Expression:
	Expression1
	Expression + Expression1
	Expression - Expression1

Expresion1:
	Expression2
	Expression1 * Expression2
	Expression1 / Expression2

Expression2:
	Expression3
	 - Expression2

Expression3:
	NUMBER
	NameRef
	( Expression )

NameRef:
	TOKEN
	TOKEN ( ArgList )  # built-in function
	TOKEN [ ArgList ]
	TOKEN . NameRef
	TOKEN [ ArgList ] . NameRef

ArgList:
	Expression
	Expression , ArgList
*/

