using System.Collections;
using TMPro;
using UnityEngine;

namespace CyberWoodenFish.game
{
    public class MainGame : MonoBehaviour
    {

        private int _combo;

        private int _gongde;
        
        [SerializeField] private TMP_Text comboText;
        
        [SerializeField] private TMP_Text gongdeText;

        [SerializeField] private GameObject subText;

        [SerializeField] private AudioSource sb;

        [SerializeField] private ParticleSystem pondParticleSystem;

        [SerializeField, Min(0f)] private float maxHitDistance = 100f;

        [SerializeField] private LayerMask hitLayers = ~0;
        
        private const float MoveSpeed = 200f; // Speed of the text movement
        
        private const float Lifetime = 0.5f; // How long the text stays visible
        
        private const float ComboHideDelay = 1f;

        private const int EffectPoolSize = 10;

        private float _lastComboUpdateTime;

        private readonly PooledClickEffect[] _effectPool = new PooledClickEffect[EffectPoolSize];

        private Camera _gameplayCamera;
        
        // Start is called once before the first execution of Update after the MonoBehaviour is created
        private void Start()
        { 
            _gongde = PlayerPrefs.GetInt("gongde"); // default 0
            _combo = 0;
            _lastComboUpdateTime = Time.time;

            _gameplayCamera = Camera.main;
            InitializeEffectPool();
        }

        // Update is called once per frame
        private void Update()
        {
            
            if (Input.GetMouseButtonDown(0) && TryHitTarget(Input.mousePosition))
            {
                _gongde -= 1;
                _combo += 1;
                _lastComboUpdateTime = Time.time;
                sb.Play();
                TryPlayPooledEffect();
            }
            
            // Update gongde text
            if (_gongde.ToString() != GetNumber(gongdeText))
            {
                PlayerPrefs.SetInt("gongde", _gongde);
                SetNumber(gongdeText, "当前功德", _gongde);
            }

            // Hide combo text if no update in the last second
            if (Time.time - _lastComboUpdateTime >= ComboHideDelay)
            {
                comboText.gameObject.SetActive(false);
            }
            else
            {
                // Show combo text if it's not visible and _combo is not zero
                if (_combo <= 0) return;
                comboText.gameObject.SetActive(true);
                if (_combo.ToString() != GetNumber(comboText))
                {
                    SetNumber(comboText, "Combo", _combo);
                }
            }
        }

        private bool TryHitTarget(Vector3 screenPosition)
        {
            var ray = _gameplayCamera.ScreenPointToRay(screenPosition);
            if (!Physics.Raycast(ray, out var hit, maxHitDistance, hitLayers, QueryTriggerInteraction.Ignore))
            {
                return false;
            }

            var target = hit.collider.GetComponentInParent<TetheredTarget>();
            if (target == null) return false;

            target.ApplyHit(hit.point, ray.direction);
            return true;
        }

        private void InitializeEffectPool()
        {
            for (var i = 0; i < EffectPoolSize; i++)
            {
                var text = Instantiate(
                    subText,
                    subText.transform.position,
                    Quaternion.identity,
                    subText.transform.parent);
                text.SetActive(false);

                var particles = Instantiate(
                    pondParticleSystem,
                    pondParticleSystem.transform.position,
                    pondParticleSystem.transform.rotation,
                    pondParticleSystem.transform.parent);
                var main = particles.main;
                main.playOnAwake = false;
                particles.gameObject.SetActive(false);

                _effectPool[i] = new PooledClickEffect(
                    text,
                    text.GetComponent<RectTransform>(),
                    particles);
            }
        }

        private void TryPlayPooledEffect()
        {
            foreach (var effect in _effectPool)
            {
                if (effect.IsInUse) continue;

                effect.IsInUse = true;
                effect.TextTransform.anchoredPosition = effect.TextStartPosition;
                effect.Text.SetActive(true);
                effect.Particles.gameObject.SetActive(true);
                effect.Particles.Play(true);
                StartCoroutine(MoveAndReleaseEffect(effect));
                return;
            }
        }

        private IEnumerator MoveAndReleaseEffect(PooledClickEffect effect)
        {
            var elapsedTime = 0f;

            while (elapsedTime < Lifetime)
            {
                var yOffset = MoveSpeed * Time.deltaTime;
                effect.TextTransform.anchoredPosition += new Vector2(0, yOffset);
                elapsedTime += Time.deltaTime;
                yield return null;
            }

            effect.Particles.Stop(true, ParticleSystemStopBehavior.StopEmittingAndClear);
            effect.Particles.gameObject.SetActive(false);
            effect.Text.SetActive(false);
            effect.IsInUse = false;
        }

        private static string GetNumber(TMP_Text text)
        {
            return text.text.Split(": ")[1];
        }

        private static void SetNumber(TMP_Text text, string type,int number)
        {
            text.text = $"{type}: {number}";
        }

        private sealed class PooledClickEffect
        {
            public PooledClickEffect(
                GameObject text,
                RectTransform textTransform,
                ParticleSystem particles)
            {
                Text = text;
                TextTransform = textTransform;
                TextStartPosition = textTransform.anchoredPosition;
                Particles = particles;
            }

            public GameObject Text { get; }

            public RectTransform TextTransform { get; }

            public Vector2 TextStartPosition { get; }

            public ParticleSystem Particles { get; }

            public bool IsInUse { get; set; }
        }
    }
}
