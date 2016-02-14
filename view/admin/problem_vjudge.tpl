{{define "content"}}
<h1>Admin - Virtual Judge Problem Add</h1>
<form accept-charset="UTF-8" class="new_vjudge" id="new_vjudge" method="post" action="/admin/problems/vjudge/">
    <div style="margin:0;padding:0;display:inline">
      <input name="utf8" type="hidden" value="✓">
    </div>
    <select id="problem_vjudge" name="vjudge" style="margin-left:10px">
      <option value="PKU">PKU</option>
      <option value="VJ">VJ</option>
    </select>
    <div class="field">
      <label for="problem_pid">Problem Pid</label><br>
      <input id="problem_pid" name="pid" size="60" type="text">
    </div>
    <div class="actions">
      <input name="commit" type="submit" value="Submit">
    </div>
</form>
<script>
var options = {
	height: '250px',
	langType : 'en',
	items: [
        'source', '|', 'undo', 'redo', '|',
        'preview', 'code', 'cut', 'copy', 'paste', 'plainpaste', 'wordpaste', '|',
        'justifyleft', 'justifycenter', 'justifyright', 'justifyfull',
        'insertorderedlist', 'insertunorderedlist', 'subscript', 'superscript',
        'clearhtml', '|', 'fullscreen', '/', 'formatblock', 'fontname', 'fontsize', '|',
        'forecolor', 'hilitecolor', 'bold', 'italic', 'underline', 'strikethrough',
        'removeformat', '|', 'image', 'table', 'hr',
        'emoticons', 'baidumap', 'link', 'unlink', '|', 'about'
	]
}

</script>
{{end}}
