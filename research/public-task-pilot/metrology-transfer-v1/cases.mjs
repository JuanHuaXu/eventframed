// Public definitions, not invented personal scenarios or private session facts.
// Only stable unit relationships are used from the historical reference table.
const B='https://www.nist.gov/pml/special-publication-811/nist-guide-si-appendix-b-conversion-factors/nist-guide-si-appendix-b8';
const S='https://www.nist.gov/pml/special-publication-811/nist-guide-si-appendix-b-conversion-factors';
const T='https://www.nist.gov/pml/owm/si-units-temperature';
const F='https://www.nist.gov/pml/special-publication-811/nist-guide-si-footnotes';
const V='https://caps.gsfc.nasa.gov/simpson/ref/unittbl.pdf';
export const groups=['international length','legacy survey length','mass conventions','gallon conventions','temperature readings','temperature intervals','area','speed','energy conventions','elapsed time','pressure','angular subdivisions'];
// group, input unit, output unit, input, scale, offset, reference output,
// source, paraphrase. Rational strings keep nonterminating quantities exact.
const rows=[
 [0,'international yard','meter','1','.9144','0',.9144,S,'How long is a yard in meters under the international length definition?'],
 [0,'international foot','meter','1','.3048','0',.3048,S,'Express a foot of ordinary international length in meters.'],
 [0,'international inch','meter','1','.0254','0',.0254,B,'What metric length in meters matches an inch, not a foot?'],
 [1,'legacy US survey foot','meter','1','1200/3937','0',.3048006096012192,S,'For historical American geodetic data, how many meters were represented by one survey foot?'],
 [1,'legacy US survey yard','meter','1','3600/3937','0',.9144018288036576,S,'Translate a historical United States survey yard into meters.'],
 [1,'legacy US survey mile','meter','1','6336000/3937','0',1609.3472186944373,S,'What length in meters belongs to the old American surveying mile rather than the international mile?'],
 [2,'avoirdupois pound','kilogram','1','.45359237','0',.45359237,F,'How many kilograms make up one ordinary commercial pound by avoirdupois mass?'],
 [2,'avoirdupois ounce','kilogram','1','.028349523125','0',.028349523125,B,'Convert an ordinary non-troy ounce of mass into kilograms.'],
 [2,'troy ounce','kilogram','1','.0311034768','0',.0311034768,B,'How much mass in kilograms is a precious-metals troy ounce, rather than an everyday ounce?'],
 [3,'US liquid gallon','liter','1','3.785411784','0',3.785411784,V,'Give the number of liters in the American wet gallon used for liquid volume.'],
 [3,'US dry gallon','liter','1','4.40488377086','0',4.40488377086,V,'Give the liter equivalent of the American dry gallon, not the liquid gallon.'],
 [3,'British Imperial gallon','liter','1','4.54609','0',4.54609,B,'What is the volume in liters of a British gallon under the Imperial convention?'],
 [4,'Fahrenheit reading','Celsius reading','32','5/9','-160/9',0,T,'A thermometer reads thirty-two degrees Fahrenheit; what Celsius reading matches it?'],
 [4,'Celsius reading','kelvin reading','0','1','273.15',273.15,T,'Translate a zero Celsius thermometer reading to the absolute kelvin scale.'],
 [4,'kelvin reading','Fahrenheit reading','273.15','9/5','-459.67',32,T,'What Fahrenheit thermometer reading corresponds to 273.15 kelvin?'],
 [5,'Fahrenheit interval','Celsius interval','1','5/9','0',5/9,B,'For a temperature rise of one degree Fahrenheit, how large is the change in degrees Celsius?'],
 [5,'Celsius interval','kelvin interval','1','1','0',1,B,'How many kelvin of temperature change equal a one-degree Celsius increase?'],
 [5,'kelvin interval','Fahrenheit interval','1','9/5','0',1.8,T,'A temperature difference is one kelvin; express that difference in Fahrenheit degrees without applying a reading offset.'],
 [6,'square inch','square meter','1','.00064516','0',.00064516,B,'What area in square meters corresponds to an inch by an inch?'],
 [6,'square foot','square meter','1','.09290304','0',.09290304,B,'Convert the area of a one-foot-by-one-foot square to square meters.'],
 [6,'square yard','square meter','1','.83612736','0',.83612736,B,'How many square meters cover a square measuring one yard on each side?'],
 [7,'mile per hour','meter per second','1','.44704','0',.44704,B,'Express a speed of one mph in meters per second.'],
 [7,'knot','meter per second','1','463/900','0',463/900,F,'How many meters per second equal one nautical mile per hour, called a knot?'],
 [7,'foot per second','meter per second','1','.3048','0',.3048,B,'Translate a speed of one foot each second into meters per second.'],
 [8,'International Table calorie','joule','1','4.1868','0',4.1868,F,'What energy in joules belongs to one calorie under the steam-table International Table convention?'],
 [8,'thermochemical calorie','joule','1','4.184','0',4.184,F,'Express a single chemistry thermochemical calorie as energy in joules, not a food kilocalorie.'],
 [8,'kilowatt hour','joule','1','3600000','0',3600000,B,'How many joules of electrical energy are in one kWh?'],
 [9,'minute of elapsed time','second of elapsed time','1','60','0',60,B,'How many seconds elapse during one clock minute rather than an angular minute?'],
 [9,'hour of elapsed time','second of elapsed time','1','3600','0',3600,B,'Express a sixty-minute clock hour as elapsed seconds.'],
 [9,'day of elapsed time','second of elapsed time','1','86400','0',86400,F,'How many elapsed seconds are in a conventional twenty-four-hour day?'],
 [10,'standard atmosphere','pascal','1','101325','0',101325,S,'What pressure in pascals is the standard atmosphere unit, not a local weather reading?'],
 [10,'bar of pressure','pascal','1','100000','0',100000,B,'Express a pressure of one bar in pascals.'],
 [10,'pound force per square inch','pascal','1','8896443230521/1290320000','0',6894.757293168361,F,'What pressure in pascals corresponds to one psi using standard gravity?'],
 [11,'arcminute','angular degree','1','1/60','0',1/60,V,'What fraction of an angular degree is a minute of arc, not a clock minute?'],
 [11,'arcsecond','angular degree','1','1/3600','0',1/3600,V,'What fraction of an angular degree is a second of arc, not elapsed time?'],
 [11,'gon','angular degree','1','.9','0',.9,B,'How many angular degrees correspond to one gradian, also called a gon?'],
];
export const cases=rows.map(([g,from,to,input,scale,offset,reference,source,paraphrase])=>({group:groups[g],split:g%2?'confirmation':'design',from,to,input,scale,offset,reference,source,paraphrase}));
