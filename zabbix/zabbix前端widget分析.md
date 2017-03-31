<!-- toc -->

# zabbix 前端widget分析
## 类层次关系
class CAreaMap extends CTag
class CColHeader extends CTag
class CCol extends CTag
class CComboBox extends CTag
class CListBox extends CComboBox
class CComboItem extends CTag
class CDiv extends CTag
    class CTabView extends CDiv
    class CUiWidget extends CDiv
    class CWarning extends CDiv
    class CClock extends CDiv
    class CColor extends CDiv
    class CColorCell extends CDiv
class CFilter extends CTag
class CForm extends CTag
class CIFrame extends CTag
class CImg extends CTag
class CInput extends CTag
    class CTextBox extends CInput
    class CNumericBox extends CInput
    class CPassBox extends CInput
    class CCheckBox extends CInput
        class CVisibilityBox extends CCheckBox
    class CFile extends CInput
class CLabel extends CTag
class CLink extends CTag
class CList extends CTag
    class CFormList extends CList
    class CHorList extends CList
    class CRadioButtonList extends CList
    class CSeverity extends CList
class CListItem extends CTag
class CMultiSelect extends CTag
class CParam extends CTag
class CPre extends CTag
class CRow extends CTag
class CRowHeader extends CTag
class CSpan extends CTag
    class CIcon extends CSpan
class CSup extends CTag
class CTable extends CTag
    class CTableInfo extends CTable
    class CTriggersInfo extends CTable
class CTextArea extends CTag

class CTag extends CObject
class CActionButtonList extends CObject
class CJsScript extends CObject
class CObject
class CTweenBox
class CVar
class CImageTextTable

class CButton extends CTag implements CButtonInterface
class CButtonCancel extends CButton
class CSubmit extends CButton
class CButtonQMessage extends CSubmit
class CButtonDelete extends CButtonQMessage

class CSimpleButton extends CTag implements CButtonInterface
class CRedirectButton extends CSimpleButton
class CSubmitButton extends CSimpleButton

class CCollapsibleUiWidget extends CUiWidget

## /include/classes/html/CObject.php 文件分析
1. CObject是其它空间的基类， 内部封装了$items属性，定义了些通用的方法实现；
2. toString(): 利用implode('', $this->items)方法，将$items属性的所有内容连接起来，形成一个字符串，并返回；
3. show(): echo $this->toString()返回的内容；
4. addItem($value): 根据$value的类型，分别执行不同的添加动作：
   1. object: 调用unpack_object($value)函数，将$value对象转换为字符串后压入$items数组；
   2. string: 直接压入$items数组；
   3. array: 对其中的每个元素，递归执行addItem();
5. unpack_object(&$item): 将$item转换为字符串；如果$item是object，这调用它的toString()方法，转换为字符串；

## /include/classes/html/CTag.php 文件分析
1. CTag继承自CObject， 表示一个html element tag， 有tagname, attributes, items等属性; 注意：构造函数的body属性通过$this->addItem($body)的形式，添加到了继承的CObject的items属性；
2. startToString(): 创建html tag的开始部分<tagName att=val>;
3. endToString(): 创建html tag的结束部分</tagName>;
4. toString(): 依次调用startToString(), bodyToString(), endToString()，将各结果连接成字符串返回；
5. addItem($value): 调用parent::addItem($value)，创建HTML tag的字符串内容；
6. setName(), getName(): 添加和获取tag的name属性；
7. getAttribute(), setAttribute, removeAttribute：获取、设置和删除tag的html属性；setAttribute($name, $value)时，$value可以是object、array和字符串；
8. addClass($class): 为tag添加$class类；
9. addAction($name, $value): 实际上是设置tag的attribute；
10. setHint(): 为tag添加一个hint box，当onMouseover或onClick的时候显示；
11. setMenuPopup(), onChange, onClick, onMouseover, onMouseout: 添加一些事件处理函数， 参数$script为字符串；
12. getForm($method='post', $action=null, $enctype=null): 创建一个Form，然后把当前html tag添加进去；

## /include/classes/html/widget/CWidget.php 文件分析
1. CWidget是zabbix前端界面的一个显示单元，controllers可以控制将它最大化，最小化等；
2. 封装了$title, $controllers, $body属性；
3. addItem($items): 将$items添加到内部的$body属性；
4. get(): 调用createTopHeader()生成$widget数组，返回[$widget, $this->body];
5. show(): 调用toString, 进而调用get()，获取字符串表示，然后echo；
6. createTopHader(): 创建$title和$controllers的DIV元素；

## /include/classes/html/CButtonCancel.php 文件分析
1. Cancel类型的button，用户点击后执行$action;
2. 原理是使用CUrl对象的getUrl()方法，获得$url，然后使用js的redirect($url)方法条状；

## CButtonQMessage 类分析
1. 继承自CSubmit;
2. 提示用户是否进行操作；

1. CCol: 继承自 CTag, 表示单元格td；
2. CColHeader: 继承自CTag， 表示标题单元格td；
3. CUiWidget: 继承自CDiv, 表示界面上可展开、可最大化、可移动的显示单元；
4. CCollapsibleUiWidget: 继承自CUiWidget, 表示可collapsed或expand的widget；
5. CColor, CColorCell: Color picker和color单元格；
6. CComboxBox: 表示下拉列表， 支持下拉列表组；构造函数的$value为选中的item；
   1. addItems(array $items): 如果$item与$this->$value一致，则创建CComboItem时，其$selected值为true；
7. CFilter: 表示页面的Filter部分，包含一个Form，Filter和Reset buttons；
   1. addColumn(): 向Filter添加column，每个column里面可以有多个控件；
8. CFormList: 继承自CList; 表示Form List，有addRow方法；
9. CHorList: 创建一个unorder horizontal list;
10. CImageTextTable: 绘制一个包含text或images的table；
11. CTabView: 创建一个tab view，包含多个table；