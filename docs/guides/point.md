---
layout: "wallarm"
page_title: "Wallarm: point"
description: |-
  Provides examples of the point argument.
---

# point

`point` - (Required) Request parts to apply the rules to. The full list of possible values is available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).
|     POINT      |POSSIBLE VALUES|
|----------------|---------------|
|`action_ext`    |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
|`action_name`   |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
|`get`           | Arbitrary GET parameter name.|
|`get_all`       |`array`, `array_all`, `array_default`, `base64`, `gzip`, `json_doc`, `xml`, `hash`, `hash_all`, `hash_default`, `hash_name`, `htmljs`, `pollution`|
|`get_default`   |`array`, `array_all`, `array_default`, `base64`, `gzip`, `json_doc`, `xml`, `hash`, `hash_all`, `hash_default`, `hash_name`, `htmljs`, `pollution`|
|`get_name`      |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
|`header`        | Arbitrary HEADER parameter name.|
|`header_all`    |`array`, `array_all`, `array_default`, `base64`, `cookie`, `cookie_all`, `cookie_default`, `cookie_name`, `gzip`, `json_doc`, `xml`, `hash`, `htmljs`, `pollution`|
|`header_default`|`array`, `array_all`, `array_default`, `base64`, `cookie`, `cookie_all`, `cookie_default`, `cookie_name`, `gzip`, `json_doc`, `xml`, `hash`, `htmljs`, `pollution`|
|`path`          | Integer value (>= 0) indicating the number of the element in the path array. |
|`path_all`      |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
|`path_default`  |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`|
|`post`          |`base64`, `form_urlencoded`, `form_urlencoded_all`, `form_urlencoded_default`, `form_urlencoded_name`, `grpc`, `grpc_all`, `grpc_default`, `gzip`, `htmljs`, `json_doc`, `multipart`, `multipart_all`, `multipart_default`, `multipart_name`, `xml`|
|`uri`           |`base64`, `gzip`, `json_doc`, `xml`,`htmljs`, `percent`|
|`json_doc`   |`array`, `array_all`, `array_default`, `hash`, `hash_all`, `hash_default`, `hash_name`, `json_array`, `json_array_all`, `json_array_default`, `json_obj`, `json_obj_all`, `json_obj_default`, `json_obj_name`|
|`instance`      | Integer ID of the application the request was sent to. |

Examples:

1. Simple form using the default content type `application/x-www-form-urlencoded`:
```
p1=1&p2[a]=2&p2[b]=3&p3[]=4&p3[]=5&p4=6&p4=7
```
may contain the points:
* `point = [["post"], ["form_urlencoded", "p1"]]` matches `1`
* `point = [["post"], ["form_urlencoded", "p2"], ["hash", "a"]]` matches `2`
* `point = [["post"], ["form_urlencoded", "p2"], ["hash", "b"]]` matches `3`
* `point = [["post"], ["form_urlencoded", "p3"], ["array", 0]]` matches `4`
* `point = [["post"], ["form_urlencoded", "p3"], ["array", 1]]` matches `5`
* `point = [["post"], ["form_urlencoded", "p4"], ["array", 0]]` matches `6`
* `point = [["post"], ["form_urlencoded", "p4"], ["array", 1]]` matches `7`
* `point = [["post"], ["form_urlencoded", "p4"], ["pollution"]]` matches `6,7`

2. JSON structure:
```json
{
"p1":"value",
"p2":[
   "v1",
   "v2"
],
"p3":{
   "somekey":"somevalue"
}
}
```
may contain the points:
* `point = [["post"],["json_doc"],["hash", "p1"]]` matches `value`
* `point = [["post"],["json_doc"],["hash", "p2"]["array", 0]]` matches `v1`
* `point = [["post"],["json_doc"],["hash", "p2"]["array", 1]]` matches `v2`
* `point = [["post"],["json_doc"],["hash", "p3"]["hash", "somekey"]]` matches `somevalue`

3. GET parameters `/?q=some+text&check=yes` may constitute the points:
* `point = [["get", "q"]]` matches `value`
* `point = [["get", "check"]]` matches `yes`

4. URL `/blogs/123/index.php?q=aaa` may constitute the points:
* `point = [["uri"]]` matches `/blogs/123/index.php?q=aaa`
* `point = [["path", 0]]` matches `blogs`
* `point = [["path", 1]]` matches `123`
* `point = [["action_name"]]` matches `index`
* `point = [["action_ext"]]` matches `php`
* `point = [["get", "q"]]` matches `aaa`

5. The request:
```
GET / HTTP/1.1
Host: example.com
X-Test: aaa
X-Test: bbb
```
may contain the points:
* `point = [["header", "HOST"]]` matches `example.com`
* `point = [["header", "X-TEST"], ["array", 0]]` matches `aaa`
* `point = [["header", "X-TEST"], ["array", 1]]` matches `aaa`
* `point = [["header", "X-TEST"], ["pollution"]]` matches `aaa,bbb`


More details on how it works are available in the [Wallarm official documentation](https://docs.wallarm.com/user-guides/rules/request-processing/).
