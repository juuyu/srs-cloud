# srs-cloud
SRS management tool.


注释参数脚本
groovyScript("if(\"${_1}\".length() == 2) {return '';} else {def result=''; def params=\"${_1}\".replaceAll('[\\\\[|\\\\]|\\\\s]', '').split(',').toList();for(i = 0; i < params.size(); i++) {if(params[i]=='null'){return;}else{result+='\\n' + ' * @param ' + params[i] + ' ' + params[i]}}; return result;}", methodParameters());

日志content
groovyScript("def params = _2.collect {'['+it+' = {}]'}.join(', '); return '\"' + _1 + '() called with parameters => ' + (params.empty ? '' : params) + '\"'", methodName(), methodParameters())
日志param
groovyScript("def params = _1.collect {it}.join(', '); return (params.empty ? '' : params) ", methodParameters())
