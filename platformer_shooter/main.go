// Platformer Shooter - Go365 Challenge Day 102
// Человечек бегает, прыгает, стреляет во врагов
// 8 апреля 2026
package main
import (
	"fmt"; "image"; "image/color"; "log"; "math"; "math/rand"; "os"
	"github.com/hajimehoshi/ebiten/v2"; "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"; "github.com/hajimehoshi/ebiten/v2/vector"
)
const (
	SW=960; SH=540; GY=450; Grav=0.6; JF=-11; PS=4.0; BS=10; WW=3000
)
type GS int
const (Menu GS=iota; Play; Over)
type Player struct { X,Y,VX,VY float64; OG bool; F,HP,MHP,AM,MAM,FT,Inv int }
type Bullet struct { X,Y,VX float64; D,L int; P bool }
type Enemy struct { X,Y,VX,Sp float64; T,HP,MHP,F,FT,pMin,pMax int; A bool }
type Part struct { X,Y,VX,VY,L,ML float64; C color.RGBA; S float64 }
type Plat struct { X,Y,W,H float64 }
type Pick struct { X,Y float64; T int; A bool }
type Game struct {
	S GS; P *Player; Bul []Bullet; En []Enemy; Pt []Part
	Pl []Plat; Pi []Pick; Cam float64; Scr,Kil,Wav,Hi int; GT float64
	Ch,BI *ebiten.Image
}
func charSp() *ebiten.Image {
	m := image.NewRGBA(image.Rect(0,0,40,50))
	for y:=34;y<48;y++ { for x:=13;x<19;x++ { m.Set(x,y,color.RGBA{50,140,70,255}) }; for x:=21;x<27;x++ { m.Set(x,y,color.RGBA{50,140,70,255}) } }
	for y:=14;y<34;y++ { for x:=10;x<30;x++ { m.Set(x,y,color.RGBA{50,140,70,255}) } }
	for y:=0;y<16;y++ { for x:=10;x<30;x++ { dx:=float64(x-20); dy:=float64(y-8); if dx*dx+dy*dy<=36 { m.Set(x,y,color.RGBA{255,210,170,255}) } } }
	m.Set(17,8,color.RGBA{0,0,0,255}); m.Set(23,8,color.RGBA{0,0,0,255})
	for y:=18;y<22;y++ { for x:=30;x<40;x++ { m.Set(x,y,color.RGBA{80,80,80,255}) } }
	return ebiten.NewImageFromImage(m)
}
func enSp(w,h int, c color.RGBA, t int) *ebiten.Image {
	m := image.NewRGBA(image.Rect(0,0,w,h))
	for y:=h/4;y<h-4;y++ { for x:=w/5;x<w*4/5;x++ { m.Set(x,y,c) } }
	for y:=0;y<h/3;y++ { for x:=w/4;x<w*3/4;x++ { dx:=float64(x-w/2); dy:=float64(y-h/5); if dx*dx+dy*dy<=float64(h/5*h/5) { m.Set(x,y,c) } } }
	m.Set(w/2-4,h/5,color.RGBA{255,50,50,255}); m.Set(w/2+4,h/5,color.RGBA{255,50,50,255})
	if t==3 { for y:=h/4;y<h/2;y++ { for x:=0;x<w/5;x++ { m.Set(x,y,color.RGBA{180,180,255,200}) }; for x:=w*4/5;x<w;x++ { m.Set(x,y,color.RGBA{180,180,255,200}) } } }
	return ebiten.NewImageFromImage(m)
}
func bulSp() *ebiten.Image {
	m := image.NewRGBA(image.Rect(0,0,12,6))
	for y:=0;y<6;y++ { for x:=0;x<12;x++ { dx:=float64(x-6); dy:=float64(y-3); if dx*dx/36+dy*dy/9<=1 { m.Set(x,y,color.RGBA{255,240,80,255}) } } }
	return ebiten.NewImageFromImage(m)
}
func lr(a,b,t float64) float64 { return a+(b-a)*t }
func di(x1,y1,x2,y2 float64) float64 { dx:=x2-x1; dy:=y2-y1; return math.Sqrt(dx*dx+dy*dy) }
func NewG() *Game {
	g := &Game{S:Menu, P:&Player{X:100,Y:GY-50,HP:100,MHP:100,AM:30,MAM:30,F:1}}
	g.Ch=charSp(); g.BI=bulSp()
	g.Pl=[]Plat{{0,GY,WW,60},{200,GY-80,150,16},{500,GY-120,200,16},{850,GY-70,180,16},{1100,GY-100,150,16},{1400,GY-60,200,16},{1700,GY-130,180,16},{2000,GY-80,150,16},{2300,GY-110,200,16},{2600,GY-70,180,16}}
	g.spawnW()
	d,_ := os.ReadFile("hi_ps.txt"); fmt.Sscanf(string(d),"%d",&g.Hi)
	return g
}
func (g *Game) spawnW() {
	n := 5+g.Wav*3; for i:=0;i<n;i++ { g.spawnE() }
	for i:=0;i<3+g.Wav;i++ { g.Pi=append(g.Pi, Pick{X:float64(200+rand.Intn(WW-400)),Y:GY-16,T:rand.Intn(2),A:true}) }
}
func (g *Game) spawnE() {
	x := 400+rand.Float64()*float64(WW-600); t:=rand.Intn(4)
	hp:=20+g.Wav*5; sp:=1.0; sz:=36
	if t==1 { hp=15+g.Wav*3; sp=0.7 }
	if t==2 { hp=60+g.Wav*10; sp=0.4; sz=44 }
	if t==3 { hp=12+g.Wav*3; sp=1.5; sz=30 }
	y := GY-float64(sz); if t==3 { y=GY-120-rand.Float64()*80 }
	g.En=append(g.En, Enemy{X:x,Y:y,VX:sp,Sp:sp,T:t,HP:hp,MHP:hp,F:-1,FT:60+rand.Intn(60),pMin:int(x)-150,pMax:int(x)+150,A:true})
}
func (g *Game) Update() error {
	g.GT += 1.0/60.0
	if g.S==Menu {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx,my:=ebiten.CursorPosition()
			if float64(mx)>=SW/2-100&&float64(mx)<=SW/2+100&&float64(my)>=420&&float64(my)<=470 { g.start() }
		}
	}
	if g.S==Over {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx,my:=ebiten.CursorPosition()
			if float64(mx)>=SW/2-90&&float64(mx)<=SW/2+90 {
				if float64(my)>=380&&float64(my)<=430 { g.start() }
				if float64(my)>=445&&float64(my)<=495 { g.S=Menu }
			}
		}
	}
	for i:=len(g.Pt)-1;i>=0;i-- { p:=&g.Pt[i]; p.X+=p.VX; p.Y+=p.VY; p.VY+=0.1; p.L-=1.0/60.0; if p.L<=0 { g.Pt=append(g.Pt[:i],g.Pt[i+1:]...) } }
	if g.S==Play { g.updatePlay() }
	return nil
}
func (g *Game) updatePlay() {
	p:=g.P; p.VX=0
	if ebiten.IsKeyPressed(ebiten.KeyA)||ebiten.IsKeyPressed(ebiten.KeyArrowLeft) { p.VX=-PS; p.F=-1 }
	if ebiten.IsKeyPressed(ebiten.KeyD)||ebiten.IsKeyPressed(ebiten.KeyArrowRight) { p.VX=PS; p.F=1 }
	if (ebiten.IsKeyPressed(ebiten.KeyW)||ebiten.IsKeyPressed(ebiten.KeyArrowUp)||ebiten.IsKeyPressed(ebiten.KeySpace))&&p.OG { p.VY=JF; p.OG=false }
	p.VY+=Grav; p.X+=p.VX; g.colX(p); p.Y+=p.VY; p.OG=false; g.colY(p)
	p.X=math.Max(10,math.Min(float64(WW)-10,p.X))
	if p.Y>float64(SH)+50 { p.HP=0 }
	if p.FT>0 { p.FT-- }; if p.Inv>0 { p.Inv-- }
	if (ebiten.IsKeyPressed(ebiten.KeyJ)||ebiten.IsKeyPressed(ebiten.KeyZ)||ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft))&&p.FT<=0&&p.AM>0 { g.shoot(); p.FT=10; p.AM-- }
	// Bullets
	for i:=len(g.Bul)-1;i>=0;i-- {
		b:=&g.Bul[i]; b.X+=b.VX; b.L--
		if b.L<=0||b.X<g.Cam-50||b.X>g.Cam+SW+50 { g.Bul=append(g.Bul[:i],g.Bul[i+1:]...); continue }
		if b.P {
			for j:=range g.En {
				e:=&g.En[j]; if !e.A { continue }
				sz:=36; if e.T==2 { sz=44 }; if e.T==3 { sz=30 }
				if di(b.X,b.Y,e.X+float64(sz)/2,e.Y+float64(sz)/2)<float64(sz)/2+4 {
					e.HP-=b.D; g.parts(b.X,b.Y,5,color.RGBA{255,200,50,255}); g.Bul=append(g.Bul[:i],g.Bul[i+1:]...)
					if e.HP<=0 {
						e.A=false; g.Kil++; g.Scr+=e.T*100+50; g.parts(e.X+float64(sz)/2,e.Y+float64(sz)/2,15,color.RGBA{255,100,50,255})
						ok:=true; for _,ee:=range g.En { if ee.A { ok=false; break } }
						if ok { g.Wav++; g.spawnW() }
					}; break
				}
			}
		} else {
			if di(b.X,b.Y,p.X+20,p.Y+25)<20&&p.Inv<=0 {
				p.HP-=b.D; p.Inv=30; g.parts(p.X+20,p.Y+25,8,color.RGBA{255,80,80,255}); g.Bul=append(g.Bul[:i],g.Bul[i+1:]...)
				if p.HP<=0 { g.S=Over; g.saveHi(); g.parts(p.X+20,p.Y+25,30,color.RGBA{255,50,50,255}) }
			}
		}
	}
	// Enemies
	for i:=range g.En {
		e:=&g.En[i]; if !e.A { continue }
		sz:=36; if e.T==2 { sz=44 }; if e.T==3 { sz=30 }
		if e.T==0||e.T==2 {
			d:=di(p.X,p.Y,e.X,e.Y)
			if d<300 { if p.X>e.X { e.X+=e.Sp; e.F=1 } else { e.X-=e.Sp; e.F=-1 } } else { e.X+=e.VX; if e.X<=float64(e.pMin)||e.X>=float64(e.pMax) { e.VX*=-1; e.F*=-1 } }
			if d<float64(sz)/2+18&&p.Inv<=0 { p.HP-=5; p.Inv=15; if p.HP<=0 { g.S=Over; g.saveHi() } }
		}
		if e.T==1 {
			e.FT--; if e.FT<=0&&di(p.X,p.Y,e.X,e.Y)<400 {
				dir:=1.0; if p.X<e.X { dir=-1 }
				g.Bul=append(g.Bul,Bullet{X:e.X+float64(sz)/2,Y:e.Y+float64(sz)/3,VX:dir*5,D:8,L:80}); e.FT=90
			}
		}
		if e.T==3 {
			dx:=p.X-e.X; dy:=p.Y-e.Y; d:=math.Sqrt(dx*dx+dy*dy)
			if d>0 { e.X+=(dx/d)*e.Sp; e.Y+=(dy/d)*e.Sp*0.3 }
			e.F=1; if dx<0 { e.F=-1 }
			if e.X<50 { e.X=50; e.F=1 }; if e.X>float64(WW)-50 { e.X=float64(WW)-50; e.F=-1 }
			if d<float64(sz)/2+18&&p.Inv<=0 { p.HP-=6; p.Inv=15; if p.HP<=0 { g.S=Over; g.saveHi() } }
		}
	}
	// Picks
	for i:=range g.Pi { pu:=&g.Pi[i]; if !pu.A { continue }; if di(p.X+20,p.Y+25,pu.X+8,pu.Y)<25 { if pu.T==0 { p.HP=int(math.Min(float64(p.MHP),float64(p.HP+25))) } else { p.AM=int(math.Min(float64(p.MAM),float64(p.AM+10))) }; pu.A=false } }
	tgt:=p.X-float64(SW)/3; g.Cam=lr(g.Cam,tgt,0.1); if g.Cam<0 { g.Cam=0.0 }; if g.Cam>float64(WW)-float64(SW) { g.Cam=float64(WW)-float64(SW) }
}
func (g *Game) colX(p *Player) { pw,ph:=30.0,46.0; for _,pl:=range g.Pl { if p.X+pw>pl.X&&p.X<pl.X+pl.W&&p.Y+ph>pl.Y+4&&p.Y<pl.Y+pl.H { if p.VX>0 { p.X=pl.X-pw } else if p.VX<0 { p.X=pl.X+pl.W }; p.VX=0 } } }
func (g *Game) colY(p *Player) { pw,ph:=30.0,46.0; for _,pl:=range g.Pl { if p.X+pw>pl.X+4&&p.X<pl.X+pl.W-4&&p.Y+ph>pl.Y&&p.Y<pl.Y+pl.H { if p.VY>0 { p.Y=pl.Y-ph; p.VY=0; p.OG=true } else if p.VY<0 { p.Y=pl.Y+pl.H; p.VY=0 } } } }
func (g *Game) shoot() { p:=g.P; d:=float64(p.F); g.Bul=append(g.Bul,Bullet{X:p.X+35*d,Y:p.Y+18,VX:d*BS,D:15,L:50,P:true}); for i:=0;i<4;i++ { g.Pt=append(g.Pt,Part{X:p.X+35*d,Y:p.Y+18,VX:d*(2+rand.Float64()*3),VY:(rand.Float64()-0.5)*3,L:0.15,ML:0.15,C:color.RGBA{255,230,60,255},S:2+rand.Float64()*2}) } }
func (g *Game) parts(x,y float64, n int, c color.RGBA) { for i:=0;i<n;i++ { a:=rand.Float64()*6.28; s:=2+rand.Float64()*4; g.Pt=append(g.Pt,Part{X:x,Y:y,VX:math.Cos(a)*s,VY:math.Sin(a)*s-1,L:0.5+rand.Float64()*0.4,ML:0.9,C:c,S:2+rand.Float64()*4}) } }
func (g *Game) saveHi() { if g.Scr>g.Hi { g.Hi=g.Scr; os.WriteFile("hi_ps.txt",[]byte(fmt.Sprintf("%d",g.Hi)),0644) } }
func (g *Game) start() { g.S=Play; g.P=&Player{X:100,Y:GY-50,HP:100,MHP:100,AM:30,MAM:30,F:1}; g.Bul=[]Bullet{}; g.En=[]Enemy{}; g.Pt=[]Part{}; g.Pi=[]Pick{}; g.Scr=0; g.Kil=0; g.Wav=1; g.Cam=0.0; g.spawnW() }
func (g *Game) Draw(screen *ebiten.Image) {
	for y:=0;y<SH;y++ { t:=float64(y)/float64(SH); r:=uint8(lr(15,40,t)); gr:=uint8(lr(12,25,t)); b:=uint8(lr(40,15,t)); vector.DrawFilledRect(screen,0,float32(y),SW,1,color.RGBA{r,gr,b,255},false) }
	cx1:=float64(g.Cam)*0.2; for x:=0;x<SW;x+=3 { h:=math.Sin(float64(x+int(cx1))*0.008)*60+math.Sin(float64(x+int(cx1))*0.02)*30+150; vector.StrokeLine(screen,float32(x),float32(h),float32(x),float32(SH),1,color.RGBA{25,30,50,80},false) }
	cx2:=float64(g.Cam)*0.4; for x:=0;x<SW;x+=3 { h:=math.Sin(float64(x+int(cx2))*0.012)*40+math.Cos(float64(x+int(cx2))*0.025)*20+250; vector.StrokeLine(screen,float32(x),float32(h),float32(x),float32(SH),1,color.RGBA{30,35,45,100},false) }
	if g.S==Menu { g.drawMenu(screen); return }
	cx:=float64(g.Cam)
	for _,pl:=range g.Pl { sx:=pl.X-cx; if sx>float64(SW)+10 || sx+pl.W < -10 { continue }; vector.DrawFilledRect(screen,float32(sx),float32(pl.Y),float32(pl.W),4,color.RGBA{80,140,80,255},false); vector.DrawFilledRect(screen,float32(sx),float32(pl.Y)+4,float32(pl.W),float32(pl.H-4),color.RGBA{50,35,25,255},false) }
	for _,pu:=range g.Pi { if !pu.A { continue }; sx:=pu.X-cx; if sx < (-20) || sx>float64(SW)+20 { continue }; c:=color.RGBA{255,80,80,255}; if pu.T==1 { c=color.RGBA{255,220,60,255} }; im:=ebiten.NewImage(16,16); for y:=0;y<16;y++ { for x:=0;x<16;x++ { dx:=float64(x-8); dy:=float64(y-8); if dx*dx+dy*dy<=64 { im.Set(x,y,c) } } }; op:=&ebiten.DrawImageOptions{}; op.GeoM.Translate(sx-8,pu.Y-8); screen.DrawImage(im,op) }
	for _,e:=range g.En { if !e.A { continue }; sx:=e.X-cx; if sx < (-50) || sx>float64(SW)+50 { continue }; sz:=36; if e.T==2 { sz=44 }; if e.T==3 { sz=30 }; var bc color.RGBA; if e.T==0 { bc=color.RGBA{180,50,50,255} }; if e.T==1 { bc=color.RGBA{200,130,40,255} }; if e.T==2 { bc=color.RGBA{130,50,180,255} }; if e.T==3 { bc=color.RGBA{50,110,180,255} }; sp:=enSp(sz,sz+8,bc,e.T); op:=&ebiten.DrawImageOptions{}; if e.F==-1 { op.GeoM.Scale(-1,1); op.GeoM.Translate(sx+float64(sz),e.Y) } else { op.GeoM.Translate(sx,e.Y) }; screen.DrawImage(sp,op); if e.HP<e.MHP { ratio:=float64(e.HP)/float64(e.MHP); vector.DrawFilledRect(screen,float32(sx),float32(e.Y-6),float32(sz),4,color.RGBA{60,0,0,200},false); vector.DrawFilledRect(screen,float32(sx),float32(e.Y-6),float32(sz)*float32(ratio),4,color.RGBA{255,50,50,255},false) } }
	p:=g.P; sx:=p.X-cx; if p.Inv<=0||int(g.GT*10)%2==0 { op:=&ebiten.DrawImageOptions{}; if p.F==-1 { op.GeoM.Scale(-1,1); op.GeoM.Translate(sx+40,p.Y) } else { op.GeoM.Translate(sx,p.Y) }; screen.DrawImage(g.Ch,op) }
	for _,b:=range g.Bul { sx:=b.X-cx; op:=&ebiten.DrawImageOptions{}; op.GeoM.Translate(sx-6,b.Y-3); if !b.P { im:=ebiten.NewImage(10,6); for y:=0;y<6;y++ { for x:=0;x<10;x++ { dx:=float64(x-5); dy:=float64(y-3); if dx*dx/25+dy*dy/9<=1 { im.Set(x,y,color.RGBA{255,60,60,255}) } } }; screen.DrawImage(im,op) } else { screen.DrawImage(g.BI,op) } }
	for _,pt:=range g.Pt { sz:=int(pt.S*(pt.L/pt.ML)); if sz<1 { continue }; a:=uint8((pt.L/pt.ML)*255); c:=color.RGBA{pt.C.R,pt.C.G,pt.C.B,a}; im:=ebiten.NewImage(sz,sz); im.Fill(c); op:=&ebiten.DrawImageOptions{}; op.GeoM.Translate(pt.X-cx-float64(sz)/2,pt.Y-float64(sz)/2); screen.DrawImage(im,op) }
	g.drawHUD(screen); if g.S==Over { g.drawOver(screen) }
}
func (g *Game) drawHUD(screen *ebiten.Image) {
	vector.DrawFilledRect(screen,0,0,SW,48,color.RGBA{10,12,30,220},false); vector.StrokeLine(screen,0,48,SW,48,2,color.RGBA{80,160,100,255},false)
	r:=float64(g.P.HP)/float64(g.P.MHP); vector.DrawFilledRect(screen,15,12,150,14,color.RGBA{50,0,0,200},false); vector.DrawFilledRect(screen,15,12,150*float32(r),14,color.RGBA{50,200,70,255},false)
	ebitenutil.DebugPrintAt(screen,fmt.Sprintf("HP %d/%d",g.P.HP,g.P.MHP),18,13)
	ar:=float64(g.P.AM)/float64(g.P.MAM); vector.DrawFilledRect(screen,15,30,100,10,color.RGBA{0,0,40,200},false); vector.DrawFilledRect(screen,15,30,100*float32(ar),10,color.RGBA{255,220,60,255},false)
	ebitenutil.DebugPrintAt(screen,fmt.Sprintf("ПАТРОНЫ: %d",g.P.AM),125,30); ebitenutil.DebugPrintAt(screen,fmt.Sprintf("СЧЁТ: %d",g.Scr),260,13); ebitenutil.DebugPrintAt(screen,fmt.Sprintf("ВОЛНА: %d",g.Wav),440,13); ebitenutil.DebugPrintAt(screen,fmt.Sprintf("УБИТО: %d",g.Kil),580,13)
	ebitenutil.DebugPrintAt(screen,"WASD - бег/прыжок | ЛКМ/J - стрельба",15,SH-18)
}
func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen,"PLATFORMER SHOOTER",SW/2-140,200); ebitenutil.DebugPrintAt(screen,"Go365 Challenge - Day 102",SW/2-130,250)
	op:=&ebiten.DrawImageOptions{}; op.GeoM.Translate(SW/2-20,280); screen.DrawImage(g.Ch,op)
	vector.DrawFilledRect(screen,0,340,SW,4,color.RGBA{80,160,80,255},false)
	g.drawBtn(screen,"START",SW/2-80,420,160,50)
	ebitenutil.DebugPrintAt(screen,"WASD - beg i pryzhok",SW/2-110,500); ebitenutil.DebugPrintAt(screen,"LKM / J - strelba",SW/2-90,520)
	if g.Hi>0 { ebitenutil.DebugPrintAt(screen,fmt.Sprintf("Rekord: %d",g.Hi),SW/2-65,550) }
}
func (g *Game) drawOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen,0,SH/2-100,SW,200,color.RGBA{10,10,20,220},false)
	ebitenutil.DebugPrintAt(screen,"GAME OVER",SW/2-80,SH/2-70)
	ebitenutil.DebugPrintAt(screen,fmt.Sprintf("Schet: %d | Volna: %d | Ubito: %d",g.Scr,g.Wav,g.Kil),SW/2-140,SH/2-20)
	g.drawBtn(screen,"RESTART",SW/2-80,SH/2+20,160,45); g.drawBtn(screen,"MENU",SW/2-80,SH/2+80,160,45)
}
func (g *Game) drawBtn(screen *ebiten.Image, t string, x,y,w,h int) {
	b:=ebiten.NewImage(w,h); mx,my:=ebiten.CursorPosition(); hv:=float64(mx)>=float64(x)&&float64(mx)<=float64(x+w)&&float64(my)>=float64(y)&&float64(my)<=float64(y+h)
	if hv { vector.DrawFilledRect(b,0,0,float32(w),float32(h),color.RGBA{50,80,120,255},false) } else { vector.DrawFilledRect(b,0,0,float32(w),float32(h),color.RGBA{30,42,70,255},false) }
	br:=ebiten.NewImage(w,3); br.Fill(color.RGBA{100,180,100,255})
	o:=&ebiten.DrawImageOptions{}; o.GeoM.Translate(float64(x),float64(y)); screen.DrawImage(b,o)
	o2:=&ebiten.DrawImageOptions{}; o2.GeoM.Translate(float64(x),float64(y)); screen.DrawImage(br,o2)
	ebitenutil.DebugPrintAt(screen,t,x+25,y+h/2-10)
}
func (g *Game) Layout(ow,oh int) (int,int) { return SW,SH }
func main() { ebiten.SetWindowSize(SW,SH); ebiten.SetWindowTitle("Platformer Shooter - Go365 Day 102"); ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled); gm:=NewG(); if err:=ebiten.RunGame(gm); err!=nil { log.Fatal(err) } }
