Before you can work with the data, you need to access it as a type other than interface{}.

In this case, the JSON represents an object, so you can use the type map[string]interface{}. 
It provides access to the next level of data. The following is a way to accessfirstName:

m := f.(map[string]interface{})
fmt.Println(m["firstName"])

At this point, the top-level keys are all accessible, allowing firstName to be accessible
by name.

To programmatically walk through the resulting data from the JSON, it’s useful to
know how Go treats the data in the conversion. When the JSON is unmarshaled, the
values in JSON are converted into the following Go types:

 bool for JSON Boolean
 float64 for JSON numbers
 []interface{} for JSON arrays
 map[string]interface{} for JSON objects
 nil for JSON null
 string for JSON strings

Knowing this, you can build functionality to walk the data structure. For example, the
following listing shows functions recursively walking the parsed JSON, printing the key
names, types, and values.

func printJSON(v interface{}) {
    switch vv := v.(type) {
    case string:
        fmt.Println("is string", vv)
    case float64:
        fmt.Println("is float64", vv)
    case []interface{}:
        fmt.Println("is an array:")
        for i, u := range vv {
            fmt.Print(i, " ")
            printJSON(u)
        }
    case map[string]interface{}:
        fmt.Println("is an object:")
        for i, u := range vv {
            fmt.Print(i, " ")
            printJSON(u)
        }
    default:
        fmt.Println("Unknown type")
    }
}