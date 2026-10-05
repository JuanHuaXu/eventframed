// Own small programs, grounded in public standard semantics, not copied tests.
const root='https://tc39.es/ecma262/2025/multipage/';
export const groups=[
 ['identity','fundamental-objects.html#sec-object.is',[
  ['signed zeros','Object.is(0,-0)','boolean:false','Does the identity predicate distinguish a negative zero from a positive zero?'],
  ['not-a-number identity','Object.is(NaN,NaN)','boolean:true','Does the identity predicate recognize two not-a-number operands as the same value?'],
  ['cross-type identity','Object.is(1,"1")','boolean:false','Does the identity predicate regard an integer and its decimal text as identical?']]],
 ['strict equality','abstract-operations.html#sec-isstrictlyequal',[
  ['signed zeros','0 === -0','boolean:true','Does strict equality regard positive and negative zero as equal?'],
  ['not-a-number equality','NaN === NaN','boolean:false','Does strict equality match a not-a-number operand with itself?'],
  ['cross-type equality','1 === "1"','boolean:false','Does strict equality accept an integer compared with decimal text?']]],
 ['coercing equality','abstract-operations.html#sec-islooselyequal',[
  ['cross-type equality','1 == "1"','boolean:true','Does loose equality match an integer with its decimal text?'],
  ['null and undefined','null == undefined','boolean:true','Does loose equality match the two empty primitive values?'],
  ['empty text and zero','"" == 0','boolean:true','Does loose equality match empty text with numeric zero?']]],
 ['array inclusion','indexed-collections.html#sec-array.prototype.includes',[
  ['not-a-number search','[NaN].includes(NaN)','boolean:true','Can the array membership predicate locate a not-a-number element?'],
  ['hole search','Array(1).includes(undefined)','boolean:true','Does the array membership predicate treat an empty slot as an undefined value?'],
  ['signed-zero search','[0].includes(-0)','boolean:true','Can the array membership predicate locate negative zero in a positive-zero slot?']]],
 ['array index search','indexed-collections.html#sec-array.prototype.indexof',[
  ['not-a-number search','[NaN].indexOf(NaN)','number:-1','Which index does the first-index search report for a not-a-number element?'],
  ['hole search','Array(1).indexOf(undefined)','number:-1','Which index does the first-index search report for an undefined value in an empty slot?'],
  ['signed-zero search','[0].indexOf(-0)','number:0','Which index does first-index search return for negative zero in a positive-zero slot?']]],
 ['array predicate search','indexed-collections.html#sec-array.prototype.findindex',[
  ['hole predicate','Array(1).findIndex(x=>x===undefined)','number:0','Does first-matching-index search run a predicate on an empty slot?'],
  ['not-a-number predicate','[NaN].findIndex(Number.isNaN)','number:0','Which index does a not-a-number predicate find in a singleton list?'],
  ['no predicate match','[1,2].findIndex(x=>x>3)','number:-1','What index represents no match in predicate-based array search?']]],
 ['JSON array serialization','structured-data.html#sec-serializejsonarray',[
  ['undefined element','JSON.stringify([undefined])','string:[null]','How is an undefined array element serialized into JSON?'],
  ['not-a-number element','JSON.stringify([NaN])','string:[null]','How is a not-a-number array element serialized into JSON?'],
  ['function element','JSON.stringify([()=>0])','string:[null]','How is a function-valued array element serialized into JSON?']]],
 ['JSON object serialization','structured-data.html#sec-serializejsonobject',[
  ['undefined property','JSON.stringify({a:undefined,b:null})','string:{"b":null}','What remains when an object containing undefined and null properties is serialized?'],
  ['function property','JSON.stringify({a:()=>0,b:1})','string:{"b":1}','What remains when an object containing a function and a numeric property is serialized?'],
  ['not-a-number property','JSON.stringify({a:NaN})','string:{"a":null}','How is a not-a-number object property represented by JSON serialization?']]],
 ['JSON top-level serialization','structured-data.html#sec-json.stringify',[
  ['undefined root','JSON.stringify(undefined)','undefined','What does top-level JSON serialization return for an undefined value?'],
  ['symbol root','JSON.stringify(Symbol())','undefined','What does top-level JSON serialization return for a symbol?'],
  ['big integer root','JSON.stringify(1n)','error:TypeError','Which exception does uncustomized JSON serialization raise for a big integer?']]],
 ['strict not-a-number test','numbers-and-dates.html#sec-number.isnan',[
  ['numeric not-a-number','Number.isNaN(NaN)','boolean:true','Does the non-coercing numeric invalid-value test recognize an actual not-a-number?'],
  ['not-a-number text','Number.isNaN("NaN")','boolean:false','Does the non-coercing numeric invalid-value test accept the text spelling of not-a-number?'],
  ['undefined input','Number.isNaN(undefined)','boolean:false','Does the non-coercing numeric invalid-value test accept undefined?']]],
 ['coercing not-a-number test','global-object.html#sec-isnan-number',[
  ['not-a-number text','isNaN("NaN")','boolean:true','Does the global numeric invalid-value test flag the text spelling of not-a-number?'],
  ['empty text','isNaN("")','boolean:false','Does the global numeric invalid-value test flag empty text?'],
  ['undefined input','isNaN(undefined)','boolean:true','Does the global numeric invalid-value test flag undefined?']]],
 ['coercing finiteness test','global-object.html#sec-isfinite-number',[
  ['null input','isFinite(null)','boolean:true','Does the global finite-number test accept a null input?'],
  ['numeric text','isFinite("42")','boolean:true','Does the global finite-number test accept decimal text?'],
  ['infinite number','isFinite(Infinity)','boolean:false','Does the global finite-number test accept positive infinity?']]],
 ['number conversion','numbers-and-dates.html#sec-number-constructor',[
  ['empty text','Number("")','number:0','What number is obtained by converting empty text with the numeric constructor?'],
  ['null input','Number(null)','number:0','What number is obtained by converting a null input with the numeric constructor?'],
  ['trailing letters','Number("12px")','number:NaN','What number is obtained when the numeric constructor sees decimal digits followed by letters?']]],
 ['floating prefix parsing','global-object.html#sec-parsefloat-string',[
  ['trailing letters','parseFloat("12px")','number:12','What value does floating-prefix parsing produce from digits followed by letters?'],
  ['no numeric prefix','parseFloat("px12")','number:NaN','What value does floating-prefix parsing produce from leading letters followed by digits?'],
  ['trailing text','parseFloat("  -2.5tail")','number:-2.5','What value does floating-prefix parsing produce from a padded negative decimal followed by text?']]],
 ['integer prefix parsing','global-object.html#sec-parseint-string-radix',[
  ['binary radix','parseInt("11",2)','number:3','What integer is obtained when two one-digits are parsed in radix two?'],
  ['decimal truncation','parseInt("2.9",10)','number:2','What integer does decimal-prefix parsing return for a positive fractional spelling?'],
  ['trailing letters','parseInt("12px",10)','number:12','What integer does decimal-prefix parsing return for digits followed by letters?']]],
 ['array sorting','indexed-collections.html#sec-array.prototype.sort',[
  ['default number order','[2,10].sort().join(",")','string:10,2','In what order does default array sorting place the numbers two and ten?'],
  ['numeric comparator','[2,10].sort((a,b)=>a-b).join(",")','string:2,10','In what order does ascending numeric comparison place two and ten during array sorting?'],
  ['mutation','(()=>{const a=[2,1];a.sort();return a.join(",")})()','string:1,2','What order remains in the original list after invoking its sorting method?']]],
 ['left reduction','indexed-collections.html#sec-array.prototype.reduce',[
  ['empty no initial','[].reduce((a,b)=>a+b)','error:TypeError','Which exception occurs when left reduction has neither elements nor an initial accumulator?'],
  ['empty initial','[].reduce((a,b)=>a+b,7)','number:7','What value does left reduction return on an empty list with a supplied accumulator?'],
  ['subtraction order','[1,2,3].reduce((a,b)=>a-b)','number:-4','What result does left-to-right reduction produce when repeatedly subtracting the next item?']]],
 ['right reduction','indexed-collections.html#sec-array.prototype.reduceright',[
  ['empty no initial','[].reduceRight((a,b)=>a+b)','error:TypeError','Which exception occurs when right reduction has neither elements nor an initial accumulator?'],
  ['empty initial','[].reduceRight((a,b)=>a+b,7)','number:7','What value does right reduction return on an empty list with a supplied accumulator?'],
  ['subtraction order','[1,2,3].reduceRight((a,b)=>a-b)','number:0','What result does right-to-left reduction produce when repeatedly subtracting the next item?']]],
];
export const cases=groups.flatMap(([group,section,rows],g)=>rows.map(([label,program,expected,paraphrase])=>({group,label,program,expected,paraphrase,source:root+section,split:['fit','design','confirmation'][g%3]})));
