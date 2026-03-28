package mbus
import "fmt"
import "time"
import "github.com/tarm/serial"
import "net"
var S *serial.Port
	var (
	Omact = make(chan bool) //флаг активности взаимодействия с OM310
	)
type Dfom []uint16  //тип значений считываемых/записываемых по modbus протоколу
func Cserial(portname string, baud int, stopb byte) error {
    var er error
    c := new(serial.Config)
	c.Name = portname //"/dev/ttyUSB0" //порт по умолчанию и его настройки
	c.Baud = baud //9600 for rs232 ОМ310
	c.StopBits = stopb //2 ! for rs232 ОМ310
    //c.Parity = "N" //"N" -no parity
	c.ReadTimeout = time.Millisecond * 1000 //
    S, er = serial.OpenPort(c)
    if er != nil {
        fmt.Printf ("%s - не найден, или занят, или нет прав у пользователя \n", c.Name)
    }
	return er
}
func (p *Dfom) Getrtu(sadr, nw uint16, mba byte) (error) {
	//m-длина принятого сообщения включая КС
	var cs1, cs2 uint16
	b := make ([]byte, 128)
	b[0] = mba		//modbus адрес OM310
	b[1] = 3		//режим чтения (hold registers)
	b[2] = byte(sadr>>8)	// старший байт начального адреса
	b[3] = byte(sadr)	// младший байт начального адреса
	b[4] = byte(nw>>8)	// старший байт числа слов
	b[5] = byte(nw)		// младший байт числа слов
	//lm = int(b[5]<<1+3+2) // должно быть принято байт от ОМ310 в режиме "чтение"	
	cs1 =Crcsum(b,6)
	b[6] = byte (cs1)
	b[7] = byte (cs1>>8)
	bm := b[8:] 	//сюда принимаем ответ
	m, err := Getoms(b) //посылаем команду и получаем ответ
	if m == 0 || err != nil {
            fmt.Println(err)  
        return err
	}
	cs2 =Crcsum (bm, m-2)
	cs1 = uint16(bm[m-1])
	cs1 = (cs1<<8)^uint16(bm[m-2])
	if (cs1^cs2)!=0 {
	err = fmt.Errorf ("Ошибка КС при приёме: %d %d",cs1,cs2)
	return err	
	}
	nad := int(bm[2]) // длина блока полученных данных в байтах в режиме чтения
    ara := *p
	if len(ara) > 0 { *p = ara[:0] } //сброс в начало среза
	//*p = nil	//сброс в начало для повторного использования одного и того же имени среза
	for i := 0; i < nad ; i++ { 
		x := uint16(bm[3+i])
		x <<= 8
		x |= uint16(bm[3+i+1])
		i++
		*p = append (*p, x)
	}	 
	return nil
}
// чтение данных из ОМ310 по сетевому соединению
func (p *Dfom) Gettcp(sadr, nw uint16, conn net.Conn, mad byte) (error) {
	var m, lm int	 //lm-длина принятого сообщения включая КС
	var cs1, cs2 uint16
	b := make ([]byte, 128)
	b[0] = mad      //1		//физ. адрес OM310
	b[1] = 3		//режим чтения
	b[2] = byte(sadr>>8)	// старший байт начального адреса
	b[3] = byte(sadr)	// младший байт начального адреса
	b[4] = byte(nw>>8)	// старший байт числа слов данных
	b[5] = byte(nw)		// младший байт числа слов
	lm = int(b[5]<<1+3+2) // должно быть принято байт от ОМ310 в режиме "чтение"	
	cs1 =Crcsum(b,6)
	b[6] = byte (cs1)
	b[7] = byte (cs1>>8)
	bm := b[8:] 	//сюда принимаем ответ
	// отправляем сообщение серверу
        n, err := conn.Write(b[:8])
	if n == 0 || err != nil { 
            fmt.Println(err)  
            return err
	}
	// приём ответа
	if m, err = conn.Read(bm[:lm]); (m < 4 || err !=nil) {
		fmt.Println(err)
		return err 
	}
	cs2 =Crcsum (bm, m-2)
	cs1 = uint16(bm[m-1])
	cs1 = (cs1<<8)^uint16(bm[m-2])
	if (cs1^cs2)!=0 {
	    err = fmt.Errorf ("Ошибка КС при приёме: %d %d",cs1,cs2)
	return err	
	}
	nad := int(bm[2]) // длина блока полученных данных в байтах в режиме чтения
	*p = nil	//сброс в начало для повторного использования одного и того же имени среза
	for i := 0; i < nad ; i++ { 
		x := uint16(bm[3+i])
		x <<= 8
		x |= uint16(bm[3+i+1])
		i++
		*p = append (*p, x)
	}	 
	return nil
}
func (p *Dfom) Getrtua(sadr, nw uint16, mba byte) (error) {
	//m-длина принятого сообщения включая КС
	var cs1, cs2 uint16
	b := make ([]byte, 128)
	b[0] = mba		//физ. адрес OM310
	b[1] = 3		//режим чтения
	b[2] = byte(sadr>>8)	// старший байт начального адреса
	b[3] = byte(sadr)	// младший байт начального адреса
	b[4] = byte(nw>>8)	// старший байт числа слов
	b[5] = byte(nw)		// младший байт числа слов
	//lm = int(b[5]<<1+3+2) // должно быть принято байт от ОМ310 в режиме "чтение"	
	cs1 =Crcsum(b,6)
	b[6] = byte (cs1)
	b[7] = byte (cs1>>8)
	bm := b[8:] 	//сюда принимаем ответ
	m, err := Getoms(b) //посылаем команду и получаем ответ
	if m == 0 || err != nil {
           	fmt.Println(err)  
        return err
	}
	cs2 = Crcsum (bm, m-2)
	cs1 = uint16(bm[m-1])
	cs1 = (cs1<<8)^uint16(bm[m-2])
	if (cs1^cs2)!=0 {
	err = fmt.Errorf ("Ошибка КС при приёме: %d %d",cs1,cs2)
	return err	
	}
	nad := int(bm[2]) // длина блока полученных данных в байтах в режиме чтения
    ara := *p
	if len(ara) > 0 { *p = ara[:0] } //сброс в начало среза
	//*p = nil	//сброс в начало для повторного использования одного и того же имени среза
	for i := 0; i < nad ; i++ { 
		x := uint16(bm[3+i])
		x <<= 8
		x |= uint16(bm[3+i+1])
		i++
		*p = append (*p, x)
	}	 
	return nil
}
func Putrtu(sadr, d uint16, mba byte) (error) {
	//var m, lm int	 //lm-длина принятого сообщения включая КС
	var cs1 uint16
	b := make ([]byte, 32)
	b[0] = mba
	b[1] = 	6	// код команды "запись 1 слова"
	b[2] = byte(sadr>>8)	// старший байт начального адреса
	b[3] = byte(sadr)	// младший байт начального адреса
	b[4] = 0		// старший байт данных
	b[5] = byte(d)		// и младший 
	//lm = 8	//длина ответа в байтах в режиме "запись 1 слова"
	cs1 = Crcsum(b,6)
	b[6] = byte (cs1)
	b[7] = byte (cs1>>8)
	m, err := Getoms (b) //посылаем команду и получаем ответ
	if m == 0 || err != nil { 
            fmt.Println(err)  
            return err
	}
	return nil //при записи параметра, ОМ310 возвращает только саму команду (без данных)
}
//отправляем команду и принимаем ответ (modbus rtu ) 
func Getoms(b []byte) (int, error) {
	Omact <- true // синхронизация с ограничителем частоты обращения к ОМ310
	var er error
	var m, m1, n int
	ms := 50 //задержка  в режиме чтения (миллисекунды) //50
	bm := b[8:] 
	lm := int(b[5]<<1+3+2) //ожидаемая длина при приёме из ОМ310 +КС
	if b[1] == 6 {
		lm = 8  // 8 байт длина ответа в режиме записи +КС
	ms = 100 //
	}
	//fmt.Println("to   om310:",b[:8]) //
	if _, er = S.Write(b[:8]); er != nil {
		return 0, er
	}
	time.Sleep(time.Millisecond * time.Duration(ms)) // задержка на передачу команды из 8 байт ! //50

	if m, er = S.Read (bm[:lm]); (er !=nil) {
		return 0, er
	}
	if m != lm {
	fmt.Println("lm, m :", lm, m)
	//если приняли не всё - считываем остаток
	bm1 := bm[m:]
	m1 = lm - m
	n, er = S.Read(bm1[:m1])
	m = lm
	if n != m1 {m = 0 } // нет ничего, это сбой
	}
	return m, nil
}
// Запись данных по modbus tcp (отсылка команды)
func Puttcp(sadr, d uint16, conn net.Conn, mad byte) (error) {
	var m, lm int	 //lm-длина принятого сообщения включая КС
	var cs1, cs2 uint16
	b := make ([]byte, 64)
	b[0] = mad      //1
	b[1] = 	6	// код команды "запись 1 слова"
	b[2] = byte(sadr>>8)	// старший байт начального адреса
	b[3] = byte(sadr)	// младший байт начального адреса
	b[4] = byte(d>>8)   // старший байт данных
	b[5] = byte(d)      // и младший 
	lm = 8	//длина ответа в байтах в режиме "запись 1 слова"
	cs1 =Crcsum(b,6)
	b[6] = byte (cs1)
	b[7] = byte (cs1>>8)
	bm := b[8:] 	//сюда принимаем ответ
	// отправляем его серверу
        n, err := conn.Write(b[:8])
	if n == 0 || err != nil { 
            fmt.Println(err)  
            return err
	}
	// приём ответа
	if m, err = conn.Read(bm[:lm]); (m < 4 || err !=nil) {
		fmt.Println(err)
		return err 
	}
	cs2 =crcsum (bm, m-2)
	cs1 = uint16(bm[m-1])
	cs1 = (cs1<<8)^uint16(bm[m-2])
	if (cs1^cs2)!=0 {
	    err = fmt.Errorf ("Ошибка КС при приёме: %d %d",cs1,cs2)
	return err	
	}	 
	return nil
}
func Crcsum(b []byte, nb int) uint16 {
	var c uint16
	c=0xFFFF
	for i:= 0;i <nb; i++ {
		c=c^uint16(b[i])
		for j:=0;j<8;j++ {
			if c&1==1 {
				c=c>>1^0xA001
			}else {
				c=c>>1}
		}
	}
	return c
}
func Crctest( b []byte, nb int) uint16 {
	cs1 := Crcsum(b, nb) //!
	cs2 := uint16(b[nb+1])
	cs2 = cs2<<8|uint16(b[nb])
	return cs1^cs2
}
//func crcput(b []byte, nb int) {
//	cs1 := Crcsum(b, nb)
//	b[nb] = byte(cs1)
//	b[nb+1] = byte(cs1 >>8)
//}

