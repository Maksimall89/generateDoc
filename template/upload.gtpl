<!DOCTYPE html>
<html>
<head>
       <title>Generate doc from Grafana</title>
</head>
<body>
<h2>Форма для задания параметров генератора отчётов</h2>
   	<form method="POST">

   		<label>Project:</label><br />
   		<select name="project" required tabindex="1">
   		    {{range .Projects}}
   		        <option value="{{.Name}}">{{.Name}}</option>
   		    {{end}}
   		</select>
        <br />

        <label>Time zone:</label><br />
      	<select name="timezone" required tabindex="2">
           		    {{range .TimeZone}}
           		      <option value="{{.}}">{{.}}</option>
           		    {{end}}
           		</select>
                <br />

   		<label>Start time:</label>
   		<br />
   		<input type="text" name="timeFrom" value="{{ .Times}}"  pattern="\d+/\d+/\d+ \d+:\d+:\d+" required tabindex="3">
   		<br />

        <label>Stop time:</label>
        <br />
        <input type="text" name="timeTo" value="{{ .Times}}"  pattern="\d+/\d+/\d+ \d+:\d+:\d+" required tabindex="4">
        <br />

        <br />
        <input type="submit" value="Generate">
</form>
</body>
</html>