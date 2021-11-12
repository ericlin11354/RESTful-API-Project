# [CSC301 Assignment 2 Group 69](https://gitlab.com/csc301-assignments/a2)

### By Supanat Wangsutthitham and Eric Lin

<br><br>

# Pair Programming

While we were doing pair programming, we decided to work on the basics of the structure of the system, e.g. the data structures that we would create, what libraries we are using, etc, so that we could later on split our work and begin working independently, taking advantage of the VCS. \
Note that ultimately, when we were debugging our code together, we again were doing pair programming, alternating between our PCs.

We used Liveshare on Visual Studio Code, hosted on Supanat’s PC, to achieve such a feat as recommended by the course instructors. We would then have a person write a portion of code. Eventually once the portion is finished, we would stop writing and discussing what would be our next step, while also switching the position between the driver and the navigator.

Note that if either of us gets fatigued, we would then stop altogether, take a break, and then come back to the appointed time.

## Reflections

- Positives
  - It is faster to finish a task as now we have another person helping us once we struggle with a certain part of the code (e.g. when we try to debug or when we forgot the functions to call on)
  - It allows us to see how different people code—his or her styling, naming convention, logical process, etc.
- Negatives
  - Liveshare is sometimes buggish, causing us some time to wait for the other person to rejoin and begin the process again. Moreover, while you can also see your partner's terminal, it is somewhat buggish and not as responsive as you would expect it to be.
  - When there is a topic that the navigator fully understands but the driver knows little to none, it can be frustrating (and time consuming) to see the driver struggles, knowing that the navigator can do such things in a much faster fashion.
  - Similarly, it could leave the navigator feeling inadequate (and potentially wasting his/her time) when the driver can code flawlessly without having to stop as the navigator does not technically have to guide the driver.

# Program Design

Our application has two main objects: TimeSeries and DailyReports. This decision was quite easy to decide on as we simply create objects that support the data type the assignment has specified.

As we are following RESTful API architecture, we decided to separate our endpoints into 2:
`/api/v1/time_series` and `/api/v1/daily_reports`.

- The `/api` is to specify the following sub endpoints for making requests to the API.
- The `/v1` part is to allow for flexibility in case we would drastically change our structure, so that it would not affect users who are still using the v1 api.
- And lastly, for the objects (`/time_series` and `/daily_reports`) we are directly following the RESTful API—each endpoint represents an object the user can perform the requests call to.

Note that we did not make endpoints for individual `TimeSeries` and `DailyReports` object as it is not in the requirements of the assignment that we have to be able to get the object by id. Nonetheless, in the case where the user wants a specific `TimeSeries` or `DailyReports`, one can do so by specifying the ID of such object in the URL paremeters (documented below).

Since we are using **Golang**, we also separated them into two modules—timeSeries and dailyReports—which each also contains handlers (of that data type) for the incoming requests. We also separate the tables in the database that we use to store them. This means that if one uploads a CSV file of `DailyReports`, it will not show up in `TimeSeries`, making them completely decoupled from each other. Again, this strictly follows the RESTful API architecture as we have decided that they are different objects. We acknowledge that this could be cause some inconvenience as a user would have to add the same data (in a different format) twice, but ultimately decided that it is for the best as it would allow for further extension and the application to be future-proof.

On the bright side, as these two objects share a lot of similarity, we were able to recycle a lot of code, some through logical processes and some through helper functions.

# Documentations

> Note: we were trying to use Swagger, but due to the time constraint—ironically, even with the request for the 48 hours extension—we do not have the time to learn how to use Swagger properly. Hence, we ended up writing our documentations in this **README** instead.

For the query type parameters, do not include `""` (double-quotation mark) nor `''` (single-quotation mark) as this will render the request invalid.

One can query multiple values in for a parameter by the following: `param=value1,value2,...` \
This unfortunately implies that any values with `,` (comma) are bound to give undesired results as the application will treat the latter as another value (coupled with the fact that we have not yet support usage of quotation marks).

Any parameters included in the query other than the ones documented will also make the request invalid.

When making a POST request to the application, only CSV files are accepted; any requests with CSV files containing duplicated dates will be rejected. \
POST requests will also update the existing data in the system if such record has already been uploaded before.

### **`/api/v1/time_series`**

- **GET**

  | Parameter              | Type   | Mandatory? | Example  | Notes                                   |
  | ---------------------- | ------ | ---------- | -------- | --------------------------------------- |
  | `id`                   | query  | no         | 1        |                                         |
  | `admin2`               | query  | no         | Autauga  |                                         |
  | `province` / `state`   | query  | no         | Ontario  | Both are interchangable                 |
  | `country` / `region`   | query  | no         | Canada   | Both are interchangable                 |
  | `date` / `from` / `to` | query  | no         | 1/31/20  | mm/dd/yy                                |
  | `death` / `recovered`  | query  | no         | death    | Both are mutually exclusive<sup>1</sup> |
  | `Accept`               | header | no         | text/csv | Default to `application/json`           |

  1: To get `confirmed` TimeSeries, leave this query blank

- **POST**

  | Parameter  | Type   | Mandatory? | Example   |
  | ---------- | ------ | ---------- | --------- |
  | `FileType` | header | yes        | Confirmed |

### **`/api/v1/daily_reports`**

- **GET**

| Parameter              | Type   | Mandatory? | Example  | Notes                         |
| ---------------------- | ------ | ---------- | -------- | ----------------------------- |
| `id`                   | query  | no         | 1        |                               |
| `admin2`               | query  | no         | Autauga  |                               |
| `province` / `state`   | query  | no         | Ontario  | Both are interchangable       |
| `country` / `region`   | query  | no         | Canada   | Both are interchangable       |
| `date` / `from` / `to` | query  | no         | 1/31/20  | mm/dd/yy                      |
| `Accept`               | header | no         | text/csv | Default to `application/json` |

<u>Note:</u> Although `death`, `confirmed`, `recovered`, and `active` are not a valid query parameter (nor documented), it will not render the request invalid; it will simply be ignored.

- **POST**

| Parameter | Type   | Mandatory? | Example | Notes    |
| --------- | ------ | ---------- | ------- | -------- |
| `Date`    | header | yes        | 1/31/20 | mm/dd/yy |
